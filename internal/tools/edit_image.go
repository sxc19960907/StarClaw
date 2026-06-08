package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/images"
)

const defaultImageProviderCDNPrefix = "https://static.kocoro.ai/"

const (
	editImageURLsMin = 1
	editImageURLsMax = 4
)

type EditImageTool struct {
	client    imageProvider
	cdnPrefix string
}

func NewEditImageTool(client imageProvider, cdnPrefix string) *EditImageTool {
	if strings.TrimSpace(cdnPrefix) == "" {
		cdnPrefix = defaultImageProviderCDNPrefix
	}
	return &EditImageTool{client: client, cdnPrefix: cdnPrefix}
}

type editImageArgs struct {
	Prompt      string   `json:"prompt"`
	ImageURLs   []string `json:"image_urls"`
	Description string   `json:"description,omitempty"`
	Size        string   `json:"size,omitempty"`
	Quality     string   `json:"quality,omitempty"`
	N           int      `json:"n,omitempty"`
	Background  string   `json:"background,omitempty"`
}

func (t *EditImageTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name: "edit_image",
		Description: "Edit one or more existing images using an explicitly configured StarClaw image provider. " +
			"Provider outputs may be permanent public URLs; anyone with the returned URL may be able to view the image. " +
			"Source image URLs must already be accepted by the configured provider CDN boundary.",
		Parameters: imageCommonSchema(true),
		Required:   []string{"prompt", "image_urls", "description"},
	}
}

func (t *EditImageTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if t.client == nil {
		return agent.BusinessError("edit_image is not configured; an image provider client must be explicitly registered"), nil
	}
	var args editImageArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	prompt, res := validateImagePrompt(args.Prompt)
	if res.IsError {
		return res, nil
	}
	if res := validateImageURLs(args.ImageURLs, t.cdnPrefix); res.IsError {
		return res, nil
	}
	if res := validateImageCommon(args.Size, args.Quality, args.Background, args.N); res.IsError {
		return res, nil
	}
	out, err := t.client.Edit(ctx, images.EditRequest{
		Prompt:     prompt,
		ImageURLs:  append([]string(nil), args.ImageURLs...),
		Size:       args.Size,
		Quality:    args.Quality,
		N:          args.N,
		Background: args.Background,
	})
	if err != nil {
		return classifyEditImageErr(err, t.cdnPrefix), nil
	}
	return agent.ToolResult{Content: formatImageResult("Edited", out)}, nil
}

func (t *EditImageTool) RequiresApproval() bool { return true }

func (t *EditImageTool) IsSafeArgs(string) bool { return false }

var _ agent.SafeChecker = (*EditImageTool)(nil)

func validateImageURLs(urls []string, cdnPrefix string) agent.ToolResult {
	if len(urls) < editImageURLsMin {
		return agent.ValidationError("image_urls is required and must contain at least 1 URL")
	}
	if len(urls) > editImageURLsMax {
		return agent.ValidationError(fmt.Sprintf("image_urls has %d entries (max %d)", len(urls), editImageURLsMax))
	}
	for i, u := range urls {
		if !strings.HasPrefix(u, cdnPrefix) {
			return agent.ValidationError(fmt.Sprintf("image_urls[%d] = %q is outside the configured provider CDN boundary %s", i, u, cdnPrefix))
		}
	}
	return agent.ToolResult{}
}

func classifyEditImageErr(err error, cdnPrefix string) agent.ToolResult {
	switch {
	case errors.Is(err, images.ErrInvalidImageURL):
		return agent.BusinessError(fmt.Sprintf("edit_image: %v; rebuild the source URL pipeline and use URLs under %s", err, cdnPrefix))
	case errors.Is(err, images.ErrSourceTooLarge):
		return agent.ValidationError(fmt.Sprintf("edit_image: %v; reduce or compress the source image before retrying", err))
	case errors.Is(err, images.ErrUnauthorized):
		return agent.PermissionError(fmt.Sprintf("edit_image: %v; check the explicitly configured image provider API key", err))
	case errors.Is(err, images.ErrEndpointNotFound):
		return agent.BusinessError(fmt.Sprintf("edit_image: %v; provider endpoint does not expose image editing", err))
	case errors.Is(err, images.ErrBadRequest), errors.Is(err, images.ErrRequestTooLarge):
		return agent.ValidationError(fmt.Sprintf("edit_image: %v", err))
	case errors.Is(err, images.ErrUpstreamTimeout):
		return agent.BusinessError(fmt.Sprintf("edit_image: %v; reduce quality, n, or source count before retrying", err))
	case errors.Is(err, images.ErrContentRejected):
		return agent.BusinessError(fmt.Sprintf("edit_image: %v; revise the prompt or source image before retrying", err))
	case errors.Is(err, images.ErrServerConfig):
		return agent.BusinessError(fmt.Sprintf("edit_image: provider image editing is not configured server-side: %v", err))
	case errors.Is(err, images.ErrTransient):
		return agent.TransientError(fmt.Sprintf("edit_image: %v", err))
	default:
		return agent.ToolResult{Content: fmt.Sprintf("edit_image error: %v", err), IsError: true}
	}
}
