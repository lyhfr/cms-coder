package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/middleware"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Me returns the current user info.
func (c *Controller) Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	token := middleware.ParseBearerToken(r)
	if token == "" {
		return nil, middleware.NewAuthError("access token is required")
	}

	out, err := c.getUserClient().IntrospectSession(ctx, token)
	if err != nil {
		return nil, err
	}

	return &v1.MeRes{
		UserId:      out.UserId,
		Email:       out.Email,
		DisplayName: out.DisplayName,
		TenantId:    out.TenantId,
		SessionId:   out.SessionId,
		ExpiresAt:   out.ExpiresAt,
	}, nil
}
