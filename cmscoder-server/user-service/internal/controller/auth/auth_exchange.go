package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/ticket"
)

// Exchange exchanges a login ticket for a formal session.
func (c *Controller) Exchange(ctx context.Context, req *v1.ExchangeReq) (res *v1.ExchangeRes, err error) {
	out, err := c.ticketSvc.Exchange(ctx, ticket.ExchangeInput{
		LoginTicket:      req.LoginTicket,
		PluginInstanceId: req.PluginInstanceId,
	})
	if err != nil {
		return nil, err
	}

	return &v1.ExchangeRes{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    out.ExpiresIn,
		User: v1.User{
			UserId:      out.User.UserId,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			TenantId:    out.User.TenantId,
		},
	}, nil
}
