package model

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "cmscoder-web-server/api/model/v1"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Controller handles model API requests.
type Controller struct {
	upstreamBaseURL string
	upstreamApiKey  string
	defaultModel    string
	httpClient      *http.Client
}

// New creates a new model controller.
func New(upstreamBaseURL, upstreamApiKey, defaultModel string) *Controller {
	return &Controller{
		upstreamBaseURL: upstreamBaseURL,
		upstreamApiKey:  upstreamApiKey,
		defaultModel:    defaultModel,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ChatCompletions proxies an OpenAI-compatible chat completions request to the upstream.
func (c *Controller) ChatCompletions(ctx context.Context, req *v1.ChatCompletionsReq) (res *v1.ChatCompletionsRes, err error) {
	r := ghttp.RequestFromCtx(ctx)

	// Read raw request body.
	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Replace model with default if not specified or empty.
	if req.Model == "" {
		req.Model = c.defaultModel
	}

	// Build upstream request.
	upstreamURL := c.upstreamBaseURL + "/v1/chat/completions"
	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}

	// Forward headers.
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", "Bearer "+c.upstreamApiKey)

	// Forward Accept header for streaming.
	if r.GetHeader("Accept") == "text/event-stream" {
		upstreamReq.Header.Set("Accept", "text/event-stream")
	}

	// Execute upstream request.
	resp, err := c.httpClient.Do(upstreamReq)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	// For streaming responses, write directly to the response.
	if resp.Header.Get("Content-Type") == "text/event-stream" {
		r.Response.Header().Set("Content-Type", "text/event-stream")
		r.Response.Header().Set("Cache-Control", "no-cache")
		r.Response.Header().Set("Connection", "keep-alive")
		r.Response.Header().Set("X-Accel-Buffering", "no")
		r.Response.WriteHeader(http.StatusOK)

		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := r.Response.Writer.Write(buf[:n]); writeErr != nil {
					return nil, fmt.Errorf("failed to write stream chunk: %w", writeErr)
				}
				r.Response.Writer.Flush()
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return nil, fmt.Errorf("failed to read stream: %w", readErr)
			}
		}
		r.ExitAll()
		return nil, nil
	}

	// For non-streaming responses, read and return as-is.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	r.Response.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	r.Response.WriteStatus(resp.StatusCode)
	r.Response.Write(respBody)
	r.ExitAll()
	return nil, nil
}

// Models returns the list of available models.
func (c *Controller) Models(ctx context.Context, req *v1.ModelsReq) (res *v1.ModelsRes, err error) {
	defaultCfg := g.Cfg()
	models := defaultCfg.MustGet(ctx, "model.available").Strings()
	if len(models) == 0 && c.defaultModel != "" {
		models = []string{c.defaultModel}
	}

	entries := make([]v1.ModelEntry, 0, len(models))
	for _, m := range models {
		entries = append(entries, v1.ModelEntry{
			ID:      m,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "cmscoder",
		})
	}

	return &v1.ModelsRes{
		Object: "list",
		Data:   entries,
	}, nil
}
