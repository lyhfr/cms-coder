package v1

import "github.com/gogf/gf/v2/frame/g"

// IAMCallbackReq is the request for completing IAM OAuth callback.
type IAMCallbackReq struct {
	g.Meta `path:"/user-service/auth/iam/callback/complete" method:"post" tags:"AuthService" summary:"Complete IAM OAuth callback"`
	Code   string `json:"code" v:"required#Authorization code is required"`
	State  string `json:"state" v:"required#State parameter is required"`
}

// IAMCallbackRes is the response for completing IAM callback.
type IAMCallbackRes struct {
	LoopbackRedirectUrl string `json:"loopbackRedirectUrl"`
}
