package v1

import "github.com/gogf/gf/v2/frame/g"

// RefreshReq is the request for refreshing access token.
type RefreshReq struct {
	g.Meta       `path:"/api/auth/refresh" method:"post" tags:"AuthService" summary:"Refresh access token"`
	RefreshToken string `json:"refreshToken" v:"required#Refresh token is required"`
}

// RefreshRes is the response for refreshing access token.
type RefreshRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}
