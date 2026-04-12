package v1

import "github.com/gogf/gf/v2/frame/g"

// CallbackReq is the request for IAM OAuth callback.
type CallbackReq struct {
	g.Meta `path:"/api/auth/iam/callback" method:"get" tags:"AuthService" summary:"IAM OAuth callback endpoint"`
	Code   string `json:"code" v:"required#Authorization code is required"`
	State  string `json:"state" v:"required#State parameter is required"`
}

// CallbackRes redirects browser to local loopback callback.
type CallbackRes struct{}
