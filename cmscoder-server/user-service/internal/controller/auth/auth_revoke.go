package auth

import (
	"context"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/session"
)

// Revoke revokes a session.
func (c *Controller) Revoke(ctx context.Context, req *v1.RevokeReq) (res *v1.RevokeRes, err error) {
	err = c.sessionSvc.Revoke(ctx, session.RevokeInput{
		RefreshToken: req.RefreshToken,
		SessionId:    req.SessionId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RevokeRes{}, nil
}
