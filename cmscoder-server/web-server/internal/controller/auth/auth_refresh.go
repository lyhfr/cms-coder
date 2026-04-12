package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/clients/userclient"
)

// Refresh refreshes the access token using a refresh token.
func (c *Controller) Refresh(ctx context.Context, req *v1.RefreshReq) (res *v1.RefreshRes, err error) {
	out, err := c.getUserClient().RefreshSession(ctx, userclient.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	return &v1.RefreshRes{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    out.ExpiresIn,
	}, nil
}
