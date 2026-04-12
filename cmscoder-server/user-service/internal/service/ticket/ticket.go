package ticket

import (
	"context"
	"time"

	"cmscoder-user-service/internal/cache"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Service manages login tickets.
type Service struct {
	cache *cache.MemoryCache
}

// New creates a new login ticket service.
func New(c *cache.MemoryCache) *Service {
	return &Service{cache: c}
}

// Create creates a login ticket.
func (s *Service) Create(ctx context.Context, ticketId, loginId, pluginInstanceId string, ttl time.Duration) error {
	return s.cache.SetLoginTicket(ctx, &cache.LoginTicket{
		TicketId:         ticketId,
		LoginId:          loginId,
		PluginInstanceId: pluginInstanceId,
		ExpiresAt:        time.Now().Add(ttl),
	}, ttl)
}

// ExchangeInput is the input for exchanging login ticket.
type ExchangeInput struct {
	LoginTicket      string
	PluginInstanceId string
}

// ExchangeOutput is the output for exchanging login ticket.
type ExchangeOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	SessionId    string
	AgentType    string
	PluginInstance string
	User         User
}

// User contains user information.
type User struct {
	UserId      string
	Email       string
	DisplayName string
	TenantId    string
}

// Exchange exchanges a login ticket for session tokens.
// The ticket consumption is atomic — concurrent requests cannot double-consume.
func (s *Service) Exchange(ctx context.Context, in ExchangeInput) (*ExchangeOutput, error) {
	// 1. Atomically validate and consume the ticket.
	ticket, err := s.cache.TryConsumeTicket(ctx, in.LoginTicket, time.Now())
	if err != nil {
		return nil, gerror.Wrap(err, "failed to consume login ticket")
	}
	if ticket.PluginInstanceId != in.PluginInstanceId {
		return nil, gerror.New("plugin instance ID mismatch")
	}

	// 2. Get user session.
	userSession, err := s.cache.GetUserSessionByLoginId(ctx, ticket.LoginId)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get user session")
	}
	if userSession == nil {
		return nil, gerror.New("user session not found")
	}

	// 3. Get user profile.
	user, err := s.cache.GetUser(ctx, userSession.UserId)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get user profile")
	}

	expiresIn := int64(time.Until(userSession.ExpiresAt).Seconds())

	return &ExchangeOutput{
		AccessToken:    userSession.SessionId,
		RefreshToken:   userSession.RefreshToken,
		ExpiresIn:      expiresIn,
		SessionId:      userSession.SessionId,
		AgentType:      userSession.AgentType,
		PluginInstance: userSession.PluginInstance,
		User: User{
			UserId:      user.UserId,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			TenantId:    user.TenantId,
		},
	}, nil
}
