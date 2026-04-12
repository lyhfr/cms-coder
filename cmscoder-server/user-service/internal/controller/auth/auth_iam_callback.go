package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/iamcallback"
)

// IAMCallback completes IAM OAuth callback and returns loopback redirect URL.
func (c *Controller) IAMCallback(ctx context.Context, req *v1.IAMCallbackReq) (res *v1.IAMCallbackRes, err error) {
	out, err := c.iamCallbackSvc.Complete(ctx, iamcallback.CompleteInput{
		Code:  req.Code,
		State: req.State,
	})
	if err != nil {
		return nil, err
	}

	return &v1.IAMCallbackRes{
		LoopbackRedirectUrl: out.LoopbackRedirectUrl,
	}, nil
}
