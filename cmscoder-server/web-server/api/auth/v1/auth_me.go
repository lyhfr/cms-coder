package v1

import "github.com/gogf/gf/v2/frame/g"

// MeReq is the request for getting current user info.
type MeReq struct {
	g.Meta `path:"/api/auth/me" method:"get" tags:"AuthService" summary:"Get current user info"`
}

// MeRes is the response for current user info.
type MeRes struct {
	UserId      string `json:"userId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	TenantId    string `json:"tenantId"`
	Project     string `json:"project,omitempty"`
	SessionId   string `json:"sessionId"`
	ExpiresAt   string `json:"expiresAt"`
}
