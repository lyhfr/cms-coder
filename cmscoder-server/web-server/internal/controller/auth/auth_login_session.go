package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/clients/userclient"
)

// LoginSession creates a login session and returns the browser URL.
func (c *Controller) LoginSession(ctx context.Context, req *v1.LoginSessionReq) (res *v1.LoginSessionRes, err error) {
	out, err := c.getUserClient().CreateLoginSession(ctx, userclient.LoginSessionInput{
		LocalPort:        req.LocalPort,
		AgentType:        req.AgentType,
		PluginInstanceId: req.PluginInstanceId,
		ClientVersion:    req.ClientVersion,
	})
	if err != nil {
		return nil, err
	}

	return &v1.LoginSessionRes{
		LoginId:    out.LoginId,
		BrowserUrl: out.BrowserUrl,
		ExpiresAt:  out.ExpiresAt,
	}, nil
}
