package v1

import "github.com/gogf/gf/v2/frame/g"

// IntrospectReq is the request for introspecting a session.
type IntrospectReq struct {
	g.Meta      `path:"/user-service/auth/sessions/introspect" method:"get" tags:"AuthService" summary:"Introspect session"`
	AccessToken string `json:"accessToken" in:"query" v:"required#Access token is required"`
}

// IntrospectRes is the response for session introspection.
type IntrospectRes struct {
	UserId       string `json:"userId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	TenantId     string `json:"tenantId"`
	SessionId    string `json:"sessionId"`
	PluginSecret string `json:"pluginSecret"` // HMAC signing key for Model Token
	ExpiresAt    string `json:"expiresAt"`
}
