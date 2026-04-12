package v1

import "github.com/gogf/gf/v2/frame/g"

// RefreshReq is the request for refreshing session.
type RefreshReq struct {
	g.Meta       `path:"/user-service/auth/sessions/refresh" method:"post" tags:"AuthService" summary:"Refresh session"`
	RefreshToken string `json:"refreshToken" v:"required#Refresh token is required"`
}

// RefreshRes is the response for refreshing session.
type RefreshRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}
