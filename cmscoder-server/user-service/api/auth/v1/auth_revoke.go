package v1

import "github.com/gogf/gf/v2/frame/g"

// RevokeReq is the request for revoking a session.
type RevokeReq struct {
	g.Meta       `path:"/user-service/auth/sessions/revoke" method:"post" tags:"AuthService" summary:"Revoke session"`
	RefreshToken string `json:"refreshToken"`
	SessionId    string `json:"sessionId"`
}

// RevokeRes is the response for revoking a session.
type RevokeRes struct{}
