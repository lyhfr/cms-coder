package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/clients/userclient"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Callback handles IAM OAuth callback and redirects to local loopback.
func (c *Controller) Callback(ctx context.Context, req *v1.CallbackReq) (res *v1.CallbackRes, err error) {
	out, err := c.getUserClient().CompleteCallback(ctx, userclient.CallbackInput{
		Code:  req.Code,
		State: req.State,
	})
	if err != nil {
		return nil, err
	}

	// Redirect browser to local loopback address with login_ticket.
	ghttp.RequestFromCtx(ctx).Response.RedirectTo(out.LoopbackRedirectUrl)
	return &v1.CallbackRes{}, nil
}
