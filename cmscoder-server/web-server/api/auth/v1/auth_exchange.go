package v1

import "github.com/gogf/gf/v2/frame/g"

// ExchangeReq is the request for exchanging login ticket for formal session.
type ExchangeReq struct {
	g.Meta           `path:"/api/auth/exchange" method:"post" tags:"AuthService" summary:"Exchange login ticket for formal session"`
	LoginTicket      string `json:"loginTicket" v:"required#Login ticket is required"`
	PluginInstanceId string `json:"pluginInstanceId" v:"required#Plugin instance ID is required"`
}

// ExchangeRes is the response for exchanging login ticket.
type ExchangeRes struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresIn    int64      `json:"expiresIn"`
	ModelApiKey  string     `json:"modelApiKey"`
	User         UserInfo   `json:"user"`
}

// UserInfo contains basic user information.
type UserInfo struct {
	UserId      string `json:"userId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	TenantId    string `json:"tenantId"`
}
