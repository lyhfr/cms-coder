package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/session"
)

// Introspect validates an access token and returns session info.
func (c *Controller) Introspect(ctx context.Context, req *v1.IntrospectReq) (res *v1.IntrospectRes, err error) {
	out, err := c.sessionSvc.Introspect(ctx, session.IntrospectInput{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		return nil, err
	}

	return &v1.IntrospectRes{
		UserId:       out.UserId,
		Email:        out.UserEmail,
		DisplayName:  out.UserDisplayName,
		TenantId:     out.UserTenantId,
		SessionId:    out.SessionId,
		PluginSecret: out.PluginSecret,
		ExpiresAt:    out.ExpiresAt,
	}, nil
}
