package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/clients/userclient"
)

// Exchange exchanges a login ticket for a formal session.
func (c *Controller) Exchange(ctx context.Context, req *v1.ExchangeReq) (res *v1.ExchangeRes, err error) {
	out, err := c.getUserClient().ExchangeLoginTicket(ctx, userclient.ExchangeInput{
		LoginTicket:      req.LoginTicket,
		PluginInstanceId: req.PluginInstanceId,
	})
	if err != nil {
		return nil, err
	}

	return &v1.ExchangeRes{
		AccessToken:    out.AccessToken,
		RefreshToken:   out.RefreshToken,
		ExpiresIn:      out.ExpiresIn,
		ModelApiKey:    out.ModelApiKey,
		CompositeToken: out.CompositeToken,
		PluginSecret:   out.PluginSecret,
		User: v1.UserInfo{
			UserId:      out.User.UserId,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			TenantId:    out.User.TenantId,
		},
	}, nil
}
