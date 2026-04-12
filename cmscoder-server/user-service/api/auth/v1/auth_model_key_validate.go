package v1

import "github.com/gogf/gf/v2/frame/g"

// ModelKeyValidateReq is the request for validating a model API key.
type ModelKeyValidateReq struct {
	g.Meta      `path:"/user-service/auth/model-keys/validate" method:"post" tags:"AuthService" summary:"Validate model API key"`
	ModelApiKey string `json:"modelApiKey" v:"required#Model API key is required"`
}

// ModelKeyValidateRes is the response for validating a model API key.
type ModelKeyValidateRes struct {
	UserId         string `json:"userId"`
	SessionId      string `json:"sessionId"`
	AgentType      string `json:"agentType"`
	PluginInstance string `json:"pluginInstance"`
	ExpiresAt      string `json:"expiresAt"`
}
