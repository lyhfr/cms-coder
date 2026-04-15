package v1

import "github.com/gogf/gf/v2/frame/g"

// ModelTokenReq is the request for obtaining a Model Token.
// The request must include a valid HMAC-SHA256 signature using the plugin_secret.
type ModelTokenReq struct {
	g.Meta           `path:"/api/auth/model-token" method:"post" tags:"AuthService" summary:"Get short-lived Model Token (JWT) for model access"`
	AccessToken      string `json:"accessToken" v:"required#Access token is required"`
	Timestamp        int64  `json:"timestamp" v:"required#Timestamp is required"`
	Nonce            string `json:"nonce" v:"required#Nonce is required"`
	Signature        string `json:"signature" v:"required#HMAC signature is required"`
	PluginInstanceId string `json:"pluginInstanceId" v:"required#Plugin instance ID is required"`
}

// ModelTokenRes is the response containing the short-lived Model Token.
type ModelTokenRes struct {
	ModelToken string `json:"modelToken"` // JWT token
	ExpiresIn  int64  `json:"expiresIn"`  // Token lifetime in seconds
	TokenType  string `json:"tokenType"`  // Always "Bearer"
}
