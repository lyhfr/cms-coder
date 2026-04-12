package loginsession

import (
	"context"
	"time"

	"cmscoder-user-service/internal/cache"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"
)

const loginSessionTTL = 5 * time.Minute

// Service manages login sessions.
type Service struct {
	cache *cache.MemoryCache
}

// New creates a new login session service.
func New(c *cache.MemoryCache) *Service {
	return &Service{cache: c}
}

// CreateInput is the input for creating a login session.
type CreateInput struct {
	LocalPort        int
	AgentType        string
	PluginInstanceId string
	ClientVersion    string
}

// CreateOutput is the output of creating a login session.
type CreateOutput struct {
	LoginId    string
	BrowserUrl string
	ExpiresAt  string
}

// Create creates a login session and returns login info.
func (s *Service) Create(ctx context.Context, in CreateInput) (*CreateOutput, error) {
	loginId := guid.S()
	expiresAt := time.Now().Add(loginSessionTTL)

	session := &cache.LoginSession{
		LoginId:          loginId,
		State:            loginId,
		LocalPort:        in.LocalPort,
		AgentType:        in.AgentType,
		PluginInstanceId: in.PluginInstanceId,
		ClientVersion:    in.ClientVersion,
		Status:           "pending",
		ExpiresAt:        expiresAt,
	}

	if err := s.cache.SetLoginSession(ctx, session, loginSessionTTL); err != nil {
		return nil, gerror.Wrap(err, "failed to create login session")
	}
	if err := s.cache.SetLoginSessionByState(ctx, loginId, loginId, loginSessionTTL); err != nil {
		return nil, gerror.Wrap(err, "failed to index login session by state")
	}

	browserUrl := "/api/auth/login/" + loginId + "/authorize"

	return &CreateOutput{
		LoginId:    loginId,
		BrowserUrl: browserUrl,
		ExpiresAt:  expiresAt.Format(time.RFC3339),
	}, nil
}

// Get retrieves a login session by loginId.
func (s *Service) Get(ctx context.Context, loginId string) (*cache.LoginSession, error) {
	session, err := s.cache.GetLoginSession(ctx, loginId)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get login session")
	}
	if session == nil {
		return nil, gerror.New("login session not found")
	}
	if session.Status != "pending" {
		return nil, gerror.Newf("login session status is %s, not pending", session.Status)
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, gerror.New("login session has expired")
	}
	return session, nil
}
