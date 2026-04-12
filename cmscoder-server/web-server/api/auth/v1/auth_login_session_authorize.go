package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginSessionAuthorizeReq is the request for getting IAM authorize URL.
type LoginSessionAuthorizeReq struct {
	g.Meta `path:"/api/auth/login/{loginId}/authorize" method:"get" tags:"AuthService" summary:"Redirect browser to IAM authorize page"`
	LoginId string `json:"loginId" v:"required#Login ID is required" dc:"Login session ID"`
}

// LoginSessionAuthorizeRes redirects browser to IAM authorize page.
type LoginSessionAuthorizeRes struct{}
