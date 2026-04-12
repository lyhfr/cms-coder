package auth

import (
	"context"
	"time"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/modelkey"
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

	// Generate model API key bound to this session.
	ttl := time.Duration(out.ExpiresIn) * time.Second
	mkOut, err := c.modelKeySvc.GenerateModelKey(ctx, modelkey.GenerateInput{
		SessionId:      out.SessionId,
		UserId:         out.User.UserId,
		AgentType:      out.AgentType,
		PluginInstance: out.PluginInstance,
		TTL:            ttl,
	})
	if err != nil {
		return nil, err
	}

	// Generate composite token binding model API key and access token.
	compositeToken := c.modelKeySvc.GenerateCompositeToken(out.SessionId, out.AccessToken)

	return &v1.ExchangeRes{
		AccessToken:    out.AccessToken,
		RefreshToken:   out.RefreshToken,
		ExpiresIn:      out.ExpiresIn,
		ModelApiKey:    mkOut.ModelApiKey,
		CompositeToken: compositeToken,
		User: v1.User{
			UserId:      out.User.UserId,
			Email:       out.User.Email,
			DisplayName: out.User.DisplayName,
			TenantId:    out.User.TenantId,
		},
	}, nil
}
