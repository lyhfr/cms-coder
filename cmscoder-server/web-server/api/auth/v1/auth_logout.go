package v1

import "github.com/gogf/gf/v2/frame/g"

// LogoutReq is the request for logging out.
type LogoutReq struct {
	g.Meta       `path:"/api/auth/logout" method:"post" tags:"AuthService" summary:"Logout user session"`
	RefreshToken string `json:"refreshToken"`
	SessionId    string `json:"sessionId"`
}

// LogoutRes is the response for logout.
type LogoutRes struct{}
