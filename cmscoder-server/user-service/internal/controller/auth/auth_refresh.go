package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/session"
)

// Refresh refreshes a session.
func (c *Controller) Refresh(ctx context.Context, req *v1.RefreshReq) (res *v1.RefreshRes, err error) {
	out, err := c.sessionSvc.Refresh(ctx, session.RefreshInput{
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
