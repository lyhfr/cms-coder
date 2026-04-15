package session

import (
	"context"
	"time"

	"cmscoder-user-service/internal/cache"
	"cmscoder-user-service/internal/service/modelkey"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// Service manages user sessions.
type Service struct {
	cache       *cache.MemoryCache
	modelKeySvc *modelkey.Service
}

// New creates a new session service.
func New(c *cache.MemoryCache, modelKeySvc *modelkey.Service) *Service {
	return &Service{cache: c, modelKeySvc: modelKeySvc}
}

// RefreshInput is the input for refreshing a session.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput is the output for refreshing a session.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Refresh refreshes a session using a refresh token.
func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	// 1. Find session by refresh token.
	userSession, err := s.cache.GetUserSessionByRefreshToken(ctx, in.RefreshToken)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to find session")
	}
	if userSession == nil {
		return nil, gerror.New("refresh token not found")
	}
	if userSession.RevokedAt != nil {
		return nil, gerror.New("session has been revoked")
	}
	if time.Now().After(userSession.ExpiresAt) {
		return nil, gerror.New("session has expired")
	}

	// 2. Rotate refresh token.
	newRefreshToken := guid.S()
	newExpiresAt := time.Now().Add(refreshTokenTTL)

	if err := s.cache.RotateRefreshToken(ctx, userSession.SessionId, newRefreshToken, newExpiresAt, refreshTokenTTL); err != nil {
		return nil, gerror.Wrap(err, "failed to rotate refresh token")
	}

	expiresIn := int64(accessTokenTTL.Seconds())

	return &RefreshOutput{
		AccessToken:  userSession.SessionId,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RevokeInput is the input for revoking a session.
type RevokeInput struct {
	RefreshToken string
	SessionId    string
}

// Revoke revokes a session.
func (s *Service) Revoke(ctx context.Context, in RevokeInput) error {
	var sessionId string
	if in.SessionId != "" {
		sessionId = in.SessionId
	} else if in.RefreshToken != "" {
		userSession, err := s.cache.GetUserSessionByRefreshToken(ctx, in.RefreshToken)
		if err != nil {
			return gerror.Wrap(err, "failed to find session")
		}
		if userSession == nil {
			return gerror.New("session not found")
		}
		sessionId = userSession.SessionId
	} else {
		return gerror.New("either sessionId or refreshToken is required")
	}

	now := time.Now()
	if err := s.cache.RevokeSession(ctx, sessionId, now); err != nil {
		return err
	}

	// Revoke associated model API key.
	s.modelKeySvc.RevokeModelKey(ctx, sessionId)
	return nil
}

// IntrospectInput is the input for introspecting a session.
type IntrospectInput struct {
	AccessToken string
}

// IntrospectOutput is the output for introspecting a session.
type IntrospectOutput struct {
	SessionId       string
	UserId          string
	UserEmail       string
	UserDisplayName string
	UserTenantId    string
	PluginSecret    string // HMAC signing key for Model Token
	ExpiresAt       string
}

// Introspect validates an access token and returns session info.
func (s *Service) Introspect(ctx context.Context, in IntrospectInput) (*IntrospectOutput, error) {
	userSession, err := s.cache.GetUserSession(ctx, in.AccessToken)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get session")
	}
	if userSession == nil {
		return nil, gerror.New("session not found")
	}
	if userSession.RevokedAt != nil {
		return nil, gerror.New("session has been revoked")
	}
	if time.Now().After(userSession.ExpiresAt) {
		return nil, gerror.New("session has expired")
	}

	user, err := s.cache.GetUser(ctx, userSession.UserId)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get user")
	}

	return &IntrospectOutput{
		SessionId:       userSession.SessionId,
		UserId:          user.UserId,
		UserEmail:       user.Email,
		UserDisplayName: user.DisplayName,
		UserTenantId:    user.TenantId,
		PluginSecret:    userSession.PluginSecret,
		ExpiresAt:       userSession.ExpiresAt.Format(time.RFC3339),
	}, nil
}
