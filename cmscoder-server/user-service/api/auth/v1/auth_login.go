package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginReq is the request for creating a login session.
type LoginReq struct {
	g.Meta           `path:"/user-service/auth/login" method:"post" tags:"AuthService" summary:"Create login session"`
	LocalPort        int    `json:"localPort" v:"required|min:1024#Local port is required and must be >= 1024"`
	AgentType        string `json:"agentType" v:"required|in:claude-code,opencode#Agent type is required"`
	PluginInstanceId string `json:"pluginInstanceId" v:"required#Plugin instance ID is required"`
	ClientVersion    string `json:"clientVersion"`
}

// LoginRes is the response for creating a login session.
type LoginRes struct {
	LoginId    string `json:"loginId"`
	BrowserUrl string `json:"browserUrl"`
	ExpiresAt  string `json:"expiresAt"`
}
