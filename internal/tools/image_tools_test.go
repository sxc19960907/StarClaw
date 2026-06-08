package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/images"
)

type fakeImageProvider struct {
	genResp   *images.GenerateResponse
	editResp  *images.GenerateResponse
	err       error
	gotGen    images.GenerateRequest
	gotEdit   images.EditRequest
	genCalls  int
	editCalls int
}

func (f *fakeImageProvider) Generate(ctx context.Context, req images.GenerateRequest) (*images.GenerateResponse, error) {
	f.genCalls++
	f.gotGen = req
	return f.genResp, f.err
}

func (f *fakeImageProvider) Edit(ctx context.Context, req images.EditRequest) (*images.GenerateResponse, error) {
	f.editCalls++
	f.gotEdit = req
	return f.editResp, f.err
}

func imageResp(url string) *images.GenerateResponse {
	return &images.GenerateResponse{
		Images: []images.Image{{URL: url, ContentType: "image/png", SizeBytes: 10}},
		Model:  "test-image",
		Size:   "1024x1024",
	}
}

func TestRegisterLocalToolsDoesNotExposeProviderImageTools(t *testing.T) {
	t.Parallel()
	reg := RegisterLocalTools()
	if _, ok := reg.Get("generate_image"); ok {
		t.Fatal("generate_image should not be registered by default")
	}
	if _, ok := reg.Get("edit_image"); ok {
		t.Fatal("edit_image should not be registered by default")
	}
}

func TestRegisterImageToolsExplicitBoundary(t *testing.T) {
	t.Parallel()
	reg := agent.NewToolRegistry()
	RegisterImageTools(reg, nil, "")
	if reg.Count() != 0 {
		t.Fatalf("nil client should register no tools, got %d", reg.Count())
	}
	RegisterImageTools(nil, &fakeImageProvider{}, "")

	RegisterImageTools(reg, &fakeImageProvider{}, "https://cdn.example/")
	for _, name := range []string{"generate_image", "edit_image"} {
		if _, ok := reg.Get(name); !ok {
			t.Fatalf("%s not registered", name)
		}
	}
}

func TestGenerateImageValidationDoesNotCallProvider(t *testing.T) {
	fake := &fakeImageProvider{}
	tool := NewGenerateImageTool(fake)
	tests := []string{
		`not json`,
		`{}`,
		`{"prompt":"   "}`,
		`{"prompt":"x","size":"4096x4096"}`,
		`{"prompt":"x","quality":"ultra"}`,
		`{"prompt":"x","background":"chrome"}`,
		`{"prompt":"x","n":11}`,
		`{"prompt":"x","n":-1}`,
	}
	for _, args := range tests {
		res, err := tool.Run(context.Background(), args)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !res.IsError || res.ErrorCategory != agent.ErrCategoryValidation {
			t.Fatalf("args %s result = %#v", args, res)
		}
	}
	if fake.genCalls != 0 {
		t.Fatalf("provider should not be called, calls=%d", fake.genCalls)
	}
}

func TestGenerateImageRunePromptAndHappyPath(t *testing.T) {
	fake := &fakeImageProvider{genResp: imageResp("https://cdn.example/a.png")}
	tool := NewGenerateImageTool(fake)
	cjk := strings.Repeat("漢", imagePromptMaxLen)
	res, err := tool.Run(context.Background(), `{"prompt":"`+cjk+`","quality":"low","n":1}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %#v", res)
	}
	if fake.gotGen.Prompt != cjk || fake.gotGen.Quality != "low" || fake.gotGen.N != 1 {
		t.Fatalf("request = %+v", fake.gotGen)
	}
	if !strings.Contains(res.Content, "https://cdn.example/a.png") || !strings.Contains(res.Content, "permanent public URLs") {
		t.Fatalf("content = %s", res.Content)
	}

	over := strings.Repeat("漢", imagePromptMaxLen+1)
	res, _ = tool.Run(context.Background(), `{"prompt":"`+over+`"}`)
	if !res.IsError || !strings.Contains(res.Content, "32001") {
		t.Fatalf("over-length result = %#v", res)
	}
}

func TestGenerateImageNilClientAndSafety(t *testing.T) {
	tool := NewGenerateImageTool(nil)
	res, err := tool.Run(context.Background(), `{"prompt":"x"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || res.ErrorCategory != agent.ErrCategoryBusiness {
		t.Fatalf("result = %#v", res)
	}
	if !tool.RequiresApproval() {
		t.Fatal("generate_image should require approval")
	}
	if tool.IsSafeArgs(`{"prompt":"x"}`) {
		t.Fatal("generate_image should never be auto-safe")
	}
}

func TestGenerateImageErrorClassification(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want agent.ErrorCategory
	}{
		{"unauthorized", images.ErrUnauthorized, agent.ErrCategoryPermission},
		{"endpoint", images.ErrEndpointNotFound, agent.ErrCategoryBusiness},
		{"bad", images.ErrBadRequest, agent.ErrCategoryValidation},
		{"large", images.ErrRequestTooLarge, agent.ErrCategoryValidation},
		{"timeout", images.ErrUpstreamTimeout, agent.ErrCategoryBusiness},
		{"rejected", images.ErrContentRejected, agent.ErrCategoryBusiness},
		{"config", images.ErrServerConfig, agent.ErrCategoryBusiness},
		{"transient", images.ErrTransient, agent.ErrCategoryTransient},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tool := NewGenerateImageTool(&fakeImageProvider{err: tt.err})
			res, _ := tool.Run(context.Background(), `{"prompt":"x"}`)
			if !res.IsError || res.ErrorCategory != tt.want {
				t.Fatalf("result = %#v", res)
			}
		})
	}
	tool := NewGenerateImageTool(&fakeImageProvider{err: errors.New("boom")})
	res, _ := tool.Run(context.Background(), `{"prompt":"x"}`)
	if !res.IsError || res.ErrorCategory != "" {
		t.Fatalf("unknown result = %#v", res)
	}
}

func TestEditImageValidationDoesNotCallProvider(t *testing.T) {
	fake := &fakeImageProvider{}
	tool := NewEditImageTool(fake, "https://cdn.example/")
	tests := []string{
		`not json`,
		`{}`,
		`{"prompt":"x"}`,
		`{"prompt":"x","image_urls":[]}`,
		`{"prompt":"x","image_urls":["https://wrong.example/a.png"]}`,
		`{"prompt":"x","image_urls":["https://cdn.example/a.png","https://cdn.example/b.png","https://cdn.example/c.png","https://cdn.example/d.png","https://cdn.example/e.png"]}`,
		`{"prompt":"x","image_urls":["https://cdn.example/a.png"],"size":"bad"}`,
	}
	for _, args := range tests {
		res, err := tool.Run(context.Background(), args)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !res.IsError || res.ErrorCategory != agent.ErrCategoryValidation {
			t.Fatalf("args %s result = %#v", args, res)
		}
	}
	if fake.editCalls != 0 {
		t.Fatalf("provider should not be called, calls=%d", fake.editCalls)
	}
}

func TestEditImageHappyPathAndDefaultPrefix(t *testing.T) {
	src := defaultImageProviderCDNPrefix + "public/a.png"
	fake := &fakeImageProvider{editResp: imageResp("https://cdn.example/edited.png")}
	tool := NewEditImageTool(fake, "")
	res, err := tool.Run(context.Background(), `{"prompt":"add moon","image_urls":["`+src+`"],"quality":"medium"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %#v", res)
	}
	if fake.gotEdit.Prompt != "add moon" || fake.gotEdit.Quality != "medium" {
		t.Fatalf("request = %+v", fake.gotEdit)
	}
	if len(fake.gotEdit.ImageURLs) != 1 || fake.gotEdit.ImageURLs[0] != src {
		t.Fatalf("image urls = %+v", fake.gotEdit.ImageURLs)
	}
	if !strings.Contains(res.Content, "Edited 1 image") || !strings.Contains(res.Content, "https://cdn.example/edited.png") {
		t.Fatalf("content = %s", res.Content)
	}
}

func TestEditImageNilClientAndSafety(t *testing.T) {
	tool := NewEditImageTool(nil, "")
	res, err := tool.Run(context.Background(), `{"prompt":"x","image_urls":["`+defaultImageProviderCDNPrefix+`a.png"]}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || res.ErrorCategory != agent.ErrCategoryBusiness {
		t.Fatalf("result = %#v", res)
	}
	if !tool.RequiresApproval() {
		t.Fatal("edit_image should require approval")
	}
	if tool.IsSafeArgs(`{"prompt":"x"}`) {
		t.Fatal("edit_image should never be auto-safe")
	}
}

func TestEditImageErrorClassification(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want agent.ErrorCategory
	}{
		{"invalid url", images.ErrInvalidImageURL, agent.ErrCategoryBusiness},
		{"source too large", images.ErrSourceTooLarge, agent.ErrCategoryValidation},
		{"unauthorized", images.ErrUnauthorized, agent.ErrCategoryPermission},
		{"endpoint", images.ErrEndpointNotFound, agent.ErrCategoryBusiness},
		{"bad", images.ErrBadRequest, agent.ErrCategoryValidation},
		{"large", images.ErrRequestTooLarge, agent.ErrCategoryValidation},
		{"timeout", images.ErrUpstreamTimeout, agent.ErrCategoryBusiness},
		{"rejected", images.ErrContentRejected, agent.ErrCategoryBusiness},
		{"config", images.ErrServerConfig, agent.ErrCategoryBusiness},
		{"transient", images.ErrTransient, agent.ErrCategoryTransient},
	}
	src := defaultImageProviderCDNPrefix + "public/a.png"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tool := NewEditImageTool(&fakeImageProvider{err: tt.err}, "")
			res, _ := tool.Run(context.Background(), `{"prompt":"x","image_urls":["`+src+`"]}`)
			if !res.IsError || res.ErrorCategory != tt.want {
				t.Fatalf("result = %#v", res)
			}
		})
	}
	tool := NewEditImageTool(&fakeImageProvider{err: errors.New("boom")}, "")
	res, _ := tool.Run(context.Background(), `{"prompt":"x","image_urls":["`+src+`"]}`)
	if !res.IsError || res.ErrorCategory != "" {
		t.Fatalf("unknown result = %#v", res)
	}
}
