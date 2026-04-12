package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcache"
)

const (
	prefixLoginSession  = "cmscoder:login_session:"
	prefixLoginState    = "cmscoder:login_state:"
	prefixLoginTicket   = "cmscoder:login_ticket:"
	prefixUserSession   = "cmscoder:user_session:"
	prefixUser          = "cmscoder:user:"
	prefixUserByIamId   = "cmscoder:user_iam:"
)

// MemoryCache provides in-memory caching for auth data.
//
// TODO: Replace with Redis for production multi-instance deployment.
// The current in-memory implementation is suitable for single-instance testing.
// When migrating to Redis:
//   - Replace gcache.Cache with gredis.Client
//   - Adjust serialization/deserialization for Redis protocols
//   - Ensure TTL behavior matches current gcache semantics
//   - Remove the mutex — Redis Lua script or WATCH/MULTI/EXEC handles atomic consumption
type MemoryCache struct {
	cache *gcache.Cache
	mu    sync.Mutex // protects atomic ticket consumption
}

// New creates a new in-memory cache.
func New() *MemoryCache {
	return &MemoryCache{
		cache: gcache.New(),
	}
}

// LoginSession represents a login session in cache.
type LoginSession struct {
	LoginId          string    `json:"loginId"`
	State            string    `json:"state"`
	LocalPort        int       `json:"localPort"`
	AgentType        string    `json:"agentType"`
	PluginInstanceId string    `json:"pluginInstanceId"`
	ClientVersion    string    `json:"clientVersion"`
	Status           string    `json:"status"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

// UserSession represents a user session in cache.
type UserSession struct {
	SessionId      string    `json:"sessionId"`
	UserId         string    `json:"userId"`
	AgentType      string    `json:"agentType"`
	PluginInstance string    `json:"pluginInstance"`
	RefreshToken   string    `json:"refreshToken"`
	ExpiresAt      time.Time `json:"expiresAt"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
}

// LoginTicket represents a login ticket in cache.
type LoginTicket struct {
	TicketId         string     `json:"ticketId"`
	LoginId          string     `json:"loginId"`
	PluginInstanceId string     `json:"pluginInstanceId"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	ConsumedAt       *time.Time `json:"consumedAt,omitempty"`
}

// User represents a user in cache.
type User struct {
	UserId      string    `json:"userId"`
	IamUserId   string    `json:"iamUserId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	TenantId    string    `json:"tenantId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SetLoginSession stores a login session.
func (c *MemoryCache) SetLoginSession(ctx context.Context, s *LoginSession, ttl time.Duration) error {
	return c.cache.Set(ctx, prefixLoginSession+s.LoginId, s, ttl)
}

// GetLoginSession retrieves a login session by loginId.
func (c *MemoryCache) GetLoginSession(ctx context.Context, loginId string) (*LoginSession, error) {
	v, err := c.cache.Get(ctx, prefixLoginSession+loginId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	var s LoginSession
	if err := v.Struct(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SetLoginSessionByState indexes a login session by state.
func (c *MemoryCache) SetLoginSessionByState(ctx context.Context, state, loginId string, ttl time.Duration) error {
	return c.cache.Set(ctx, prefixLoginState+state, loginId, ttl)
}

// GetLoginSessionByState retrieves a login session by state.
func (c *MemoryCache) GetLoginSessionByState(ctx context.Context, state string) (*LoginSession, error) {
	v, err := c.cache.Get(ctx, prefixLoginState+state)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	loginId := v.String()
	return c.GetLoginSession(ctx, loginId)
}

// UpdateLoginSessionStatus updates the status of a login session.
func (c *MemoryCache) UpdateLoginSessionStatus(ctx context.Context, loginId, status string) error {
	s, err := c.GetLoginSession(ctx, loginId)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("login session %s not found", loginId)
	}
	s.Status = status
	ttl := time.Until(s.ExpiresAt)
	return c.cache.Set(ctx, prefixLoginSession+loginId, s, ttl)
}

// SetLoginTicket stores a login ticket.
func (c *MemoryCache) SetLoginTicket(ctx context.Context, t *LoginTicket, ttl time.Duration) error {
	return c.cache.Set(ctx, prefixLoginTicket+t.TicketId, t, ttl)
}

// GetLoginTicket retrieves a login ticket.
func (c *MemoryCache) GetLoginTicket(ctx context.Context, ticketId string) (*LoginTicket, error) {
	v, err := c.cache.Get(ctx, prefixLoginTicket+ticketId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	var t LoginTicket
	if err := v.Struct(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// TryConsumeTicket atomically validates and consumes a login ticket.
// Returns the ticket on success, or an error if invalid/already consumed/expired.
// This prevents concurrent double-consumption of the same ticket.
func (c *MemoryCache) TryConsumeTicket(ctx context.Context, ticketId string, consumedAt time.Time) (*LoginTicket, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, err := c.GetLoginTicket(ctx, ticketId)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("login ticket %s not found", ticketId)
	}
	if t.ConsumedAt != nil {
		return nil, fmt.Errorf("login ticket %s has already been consumed", ticketId)
	}
	if time.Now().After(t.ExpiresAt) {
		return nil, fmt.Errorf("login ticket %s has expired", ticketId)
	}

	t.ConsumedAt = &consumedAt
	ttl := time.Until(t.ExpiresAt)
	if err := c.cache.Set(ctx, prefixLoginTicket+ticketId, t, ttl); err != nil {
		return nil, err
	}
	return t, nil
}

// CreateUserSession stores a user session.
func (c *MemoryCache) CreateUserSession(ctx context.Context, s *UserSession, ttl time.Duration) error {
	if err := c.cache.Set(ctx, prefixUserSession+s.SessionId, s, ttl); err != nil {
		return err
	}
	// Index by refresh token.
	return c.cache.Set(ctx, prefixUserSession+"rt:"+s.RefreshToken, s.SessionId, ttl)
}

// GetUserSession retrieves a user session by session ID (access token).
func (c *MemoryCache) GetUserSession(ctx context.Context, sessionId string) (*UserSession, error) {
	v, err := c.cache.Get(ctx, prefixUserSession+sessionId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	var s UserSession
	if err := v.Struct(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetUserSessionByRefreshToken retrieves a user session by refresh token.
func (c *MemoryCache) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error) {
	v, err := c.cache.Get(ctx, prefixUserSession+"rt:"+refreshToken)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	sessionId := v.String()
	return c.GetUserSession(ctx, sessionId)
}

// GetUserSessionByLoginId retrieves a user session by login ID.
func (c *MemoryCache) GetUserSessionByLoginId(ctx context.Context, loginId string) (*UserSession, error) {
	// Lookup sessions indexed by login session.
	// We store a reverse index: login_session -> user_session
	v, err := c.cache.Get(ctx, prefixUserSession+"login:"+loginId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	sessionId := v.String()
	return c.GetUserSession(ctx, sessionId)
}

// SetUserSessionByLoginId indexes a user session by login ID.
func (c *MemoryCache) SetUserSessionByLoginId(ctx context.Context, loginId, sessionId string, ttl time.Duration) error {
	return c.cache.Set(ctx, prefixUserSession+"login:"+loginId, sessionId, ttl)
}

// RotateRefreshToken rotates the refresh token for a session.
func (c *MemoryCache) RotateRefreshToken(ctx context.Context, sessionId, newRefreshToken string, newExpiresAt time.Time, ttl time.Duration) error {
	s, err := c.GetUserSession(ctx, sessionId)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("user session %s not found", sessionId)
	}
	s.RefreshToken = newRefreshToken
	s.ExpiresAt = newExpiresAt

	if err := c.cache.Set(ctx, prefixUserSession+sessionId, s, ttl); err != nil {
		return err
	}
	// Delete old refresh token index to prevent replay.
	if s.RefreshToken != "" {
		c.cache.Remove(ctx, prefixUserSession+"rt:"+s.RefreshToken)
	}
	return c.cache.Set(ctx, prefixUserSession+"rt:"+newRefreshToken, sessionId, ttl)
}

// RevokeSession revokes a user session.
func (c *MemoryCache) RevokeSession(ctx context.Context, sessionId string, revokedAt time.Time) error {
	s, err := c.GetUserSession(ctx, sessionId)
	if err != nil {
		return err
	}
	if s == nil {
		return nil
	}
	s.RevokedAt = &revokedAt
	ttl := time.Until(s.ExpiresAt)
	return c.cache.Set(ctx, prefixUserSession+sessionId, s, ttl)
}

// SetUser stores a user profile.
func (c *MemoryCache) SetUser(ctx context.Context, u *User) error {
	ttl := 24 * time.Hour
	if err := c.cache.Set(ctx, prefixUser+u.UserId, u, ttl); err != nil {
		return err
	}
	return c.cache.Set(ctx, prefixUserByIamId+u.IamUserId, u.UserId, ttl)
}

// GetUser retrieves a user by userId.
func (c *MemoryCache) GetUser(ctx context.Context, userId string) (*User, error) {
	v, err := c.cache.Get(ctx, prefixUser+userId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	var u User
	if err := v.Struct(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByIamId retrieves a user by IAM user ID.
func (c *MemoryCache) GetUserByIamId(ctx context.Context, iamUserId string) (*User, error) {
	v, err := c.cache.Get(ctx, prefixUserByIamId+iamUserId)
	if err != nil {
		return nil, err
	}
	if v.IsNil() {
		return nil, nil
	}
	userId := v.String()
	return c.GetUser(ctx, userId)
}
