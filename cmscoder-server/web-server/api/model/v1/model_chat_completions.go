package v1

import "github.com/gogf/gf/v2/frame/g"

// ChatCompletionsReq is an OpenAI-compatible chat completions request.
type ChatCompletionsReq struct {
	g.Meta `path:"/api/model/v1/chat/completions" method:"post" tags:"ModelService" summary:"OpenAI-compatible chat completions"`
	Model  string `json:"model"`
}

// ChatCompletionsRes is an OpenAI-compatible chat completions response.
// The actual response is streamed or returned as raw JSON from the upstream.
type ChatCompletionsRes struct {
	g.Meta `mime:"application/json"`
}
