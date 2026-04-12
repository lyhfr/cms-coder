package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
)

// Logout revokes the user session.
func (c *Controller) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	err = c.getUserClient().RevokeSession(ctx, req.SessionId, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &v1.LogoutRes{}, nil
}
