package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/loginsession"
)

// Login creates a login session and returns browser URL.
func (c *Controller) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	out, err := c.loginSessionSvc.Create(ctx, loginsession.CreateInput{
		LocalPort:        req.LocalPort,
		AgentType:        req.AgentType,
		PluginInstanceId: req.PluginInstanceId,
		ClientVersion:    req.ClientVersion,
	})
	if err != nil {
		return nil, err
	}

	return &v1.LoginRes{
		LoginId:    out.LoginId,
		BrowserUrl: out.BrowserUrl,
		ExpiresAt:  out.ExpiresAt,
	}, nil
}
