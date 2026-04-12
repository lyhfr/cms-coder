package v1

import "github.com/gogf/gf/v2/frame/g"

// ModelsReq is the request for listing available models.
type ModelsReq struct {
	g.Meta `path:"/api/model/v1/models" method:"get" tags:"ModelService" summary:"List available models"`
}

// ModelsRes is the response for listing available models.
type ModelsRes struct {
	Object string       `json:"object"`
	Data   []ModelEntry `json:"data"`
}

// ModelEntry describes a single available model.
type ModelEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}
