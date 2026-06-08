package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/images"
)

type imageProvider interface {
	Generate(ctx context.Context, req images.GenerateRequest) (*images.GenerateResponse, error)
	Edit(ctx context.Context, req images.EditRequest) (*images.GenerateResponse, error)
}

const (
	imagePromptMaxLen = 32000
	imageNMin         = 1
	imageNMax         = 10
)

var (
	imageValidSizes      = map[string]bool{"1024x1024": true, "1024x1536": true, "1536x1024": true, "auto": true}
	imageValidQuality    = map[string]bool{"auto": true, "low": true, "medium": true, "high": true}
	imageValidBackground = map[string]bool{"transparent": true, "opaque": true, "auto": true}
)

type GenerateImageTool struct {
	client imageProvider
}

func NewGenerateImageTool(client imageProvider) *GenerateImageTool {
	return &GenerateImageTool{client: client}
}

type generateImageArgs struct {
	Prompt      string `json:"prompt"`
	Description string `json:"description,omitempty"`
	Size        string `json:"size,omitempty"`
	Quality     string `json:"quality,omitempty"`
	N           int    `json:"n,omitempty"`
	Background  string `json:"background,omitempty"`
}

func (t *GenerateImageTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name: "generate_image",
		Description: "Generate an image from a text prompt using an explicitly configured StarClaw image provider. " +
			"Provider outputs may be permanent public URLs; anyone with the returned URL may be able to view the image. " +
			"Use n=1 unless the user explicitly asks for variants. Use local imaging for metadata/resize/convert work.",
		Parameters: imageCommonSchema(false),
		Required:   []string{"prompt", "description"},
	}
}

func (t *GenerateImageTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if t.client == nil {
		return agent.BusinessError("generate_image is not configured; an image provider client must be explicitly registered"), nil
	}
	var args generateImageArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	prompt, res := validateImagePrompt(args.Prompt)
	if res.IsError {
		return res, nil
	}
	if res := validateImageCommon(args.Size, args.Quality, args.Background, args.N); res.IsError {
		return res, nil
	}
	out, err := t.client.Generate(ctx, images.GenerateRequest{
		Prompt:     prompt,
		Size:       args.Size,
		Quality:    args.Quality,
		N:          args.N,
		Background: args.Background,
	})
	if err != nil {
		return classifyGenerateImageErr(err), nil
	}
	return agent.ToolResult{Content: formatImageResult("Generated", out)}, nil
}

func (t *GenerateImageTool) RequiresApproval() bool { return true }

func (t *GenerateImageTool) IsSafeArgs(string) bool { return false }

var _ agent.SafeChecker = (*GenerateImageTool)(nil)

func imageCommonSchema(includeImageURLs bool) map[string]any {
	properties := map[string]any{
		"prompt": map[string]any{
			"type":        "string",
			"minLength":   1,
			"maxLength":   imagePromptMaxLen,
			"description": "Detailed image prompt or edit instruction.",
		},
		"description": map[string]any{
			"type":        "string",
			"description": "Brief description of why provider-backed image generation/editing is needed.",
		},
		"size": map[string]any{
			"type":        "string",
			"enum":        []string{"1024x1024", "1024x1536", "1536x1024", "auto"},
			"description": "Output image dimensions.",
		},
		"quality": map[string]any{
			"type":        "string",
			"enum":        []string{"auto", "low", "medium", "high"},
			"description": "Provider quality setting; higher can be slower and more expensive.",
		},
		"n": map[string]any{
			"type":        "integer",
			"minimum":     imageNMin,
			"maximum":     imageNMax,
			"description": "Number of images. 0 uses provider default; 1 is preferred unless variants were requested.",
		},
		"background": map[string]any{
			"type":        "string",
			"enum":        []string{"transparent", "opaque", "auto"},
			"description": "Background mode.",
		},
	}
	if includeImageURLs {
		properties["image_urls"] = map[string]any{
			"type":        "array",
			"minItems":    editImageURLsMin,
			"maxItems":    editImageURLsMax,
			"items":       map[string]any{"type": "string"},
			"description": "1-4 source image URLs accepted by the configured provider.",
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": properties,
	}
}

func validateImagePrompt(value string) (string, agent.ToolResult) {
	prompt := strings.TrimSpace(value)
	if prompt == "" {
		return "", agent.ValidationError("prompt is required")
	}
	if runes := utf8.RuneCountInString(prompt); runes > imagePromptMaxLen {
		return "", agent.ValidationError(fmt.Sprintf("prompt too long: %d chars (max %d)", runes, imagePromptMaxLen))
	}
	return prompt, agent.ToolResult{}
}

func validateImageCommon(size, quality, background string, n int) agent.ToolResult {
	if size != "" && !imageValidSizes[size] {
		return agent.ValidationError(fmt.Sprintf("invalid size %q: must be one of 1024x1024, 1024x1536, 1536x1024, auto", size))
	}
	if quality != "" && !imageValidQuality[quality] {
		return agent.ValidationError(fmt.Sprintf("invalid quality %q: must be one of auto, low, medium, high", quality))
	}
	if background != "" && !imageValidBackground[background] {
		return agent.ValidationError(fmt.Sprintf("invalid background %q: must be one of transparent, opaque, auto", background))
	}
	if n < 0 || n > imageNMax {
		return agent.ValidationError(fmt.Sprintf("invalid n=%d: must be 1..%d, or 0 for provider default", n, imageNMax))
	}
	return agent.ToolResult{}
}

func formatImageResult(verb string, res *images.GenerateResponse) string {
	if res == nil || len(res.Images) == 0 {
		return verb + " 0 image(s)."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d image(s). Provider outputs may be permanent public URLs.\n", verb, len(res.Images))
	for _, img := range res.Images {
		fmt.Fprintf(&b, "URL: %s\n", img.URL)
	}
	first := res.Images[0]
	if res.Size != "" || first.ContentType != "" || first.SizeBytes > 0 {
		fmt.Fprintf(&b, "Size: %s | Content-Type: %s | %d bytes\n", res.Size, first.ContentType, first.SizeBytes)
	}
	if res.Model != "" {
		fmt.Fprintf(&b, "Model: %s", res.Model)
	}
	return strings.TrimRight(b.String(), "\n")
}

func classifyGenerateImageErr(err error) agent.ToolResult {
	switch {
	case errors.Is(err, images.ErrUnauthorized):
		return agent.PermissionError(fmt.Sprintf("generate_image: %v; check the explicitly configured image provider API key", err))
	case errors.Is(err, images.ErrEndpointNotFound):
		return agent.BusinessError(fmt.Sprintf("generate_image: %v; provider endpoint does not expose image generation", err))
	case errors.Is(err, images.ErrBadRequest), errors.Is(err, images.ErrRequestTooLarge):
		return agent.ValidationError(fmt.Sprintf("generate_image: %v", err))
	case errors.Is(err, images.ErrUpstreamTimeout):
		return agent.BusinessError(fmt.Sprintf("generate_image: %v; reduce quality or n before retrying", err))
	case errors.Is(err, images.ErrContentRejected):
		return agent.BusinessError(fmt.Sprintf("generate_image: %v; revise the prompt before retrying", err))
	case errors.Is(err, images.ErrServerConfig):
		return agent.BusinessError(fmt.Sprintf("generate_image: provider image generation is not configured server-side: %v", err))
	case errors.Is(err, images.ErrTransient):
		return agent.TransientError(fmt.Sprintf("generate_image: %v", err))
	default:
		return agent.ToolResult{Content: fmt.Sprintf("generate_image error: %v", err), IsError: true}
	}
}
