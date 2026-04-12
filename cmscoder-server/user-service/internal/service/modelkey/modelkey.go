package modelkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"time"

	"cmscoder-user-service/internal/cache"
)

// Service manages model API keys.
type Service struct {
	cache *cache.MemoryCache
}

// New creates a new model key service.
func New(c *cache.MemoryCache) *Service {
	return &Service{cache: c}
}

// GenerateInput is the input for generating a model API key.
type GenerateInput struct {
	SessionId      string
	UserId         string
	AgentType      string
	PluginInstance string
	TTL            time.Duration
}

// GenerateOutput is the output of generating a model API key.
type GenerateOutput struct {
	ModelApiKey string
	ExpiresAt   time.Time
}

// GenerateModelKey creates a new model API key bound to the user session.
func (s *Service) GenerateModelKey(ctx context.Context, in GenerateInput) (*GenerateOutput, error) {
	// Delete any existing key for this session (rotation).
	s.cache.DeleteModelKeyBySession(ctx, in.SessionId)

	key := generateKey()
	expiresAt := time.Now().Add(in.TTL)

	mk := &cache.ModelKey{
		ModelApiKey:    key,
		UserId:         in.UserId,
		SessionId:      in.SessionId,
		AgentType:      in.AgentType,
		PluginInstance: in.PluginInstance,
		ExpiresAt:      expiresAt,
	}

	if err := s.cache.SetModelKey(ctx, mk, in.TTL); err != nil {
		return nil, fmt.Errorf("failed to store model key: %w", err)
	}

	return &GenerateOutput{
		ModelApiKey: key,
		ExpiresAt:   expiresAt,
	}, nil
}

// ValidateInput is the input for validating a model API key.
type ValidateInput struct {
	ModelApiKey string
}

// ValidateOutput is the output of validating a model API key.
type ValidateOutput struct {
	UserId         string
	SessionId      string
	AgentType      string
	PluginInstance string
	ExpiresAt      time.Time
}

// ValidateModelKey validates a model API key and returns its metadata.
func (s *Service) ValidateModelKey(ctx context.Context, in ValidateInput) (*ValidateOutput, error) {
	mk, err := s.cache.GetModelKey(ctx, in.ModelApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get model key: %w", err)
	}
	if mk == nil {
		return nil, fmt.Errorf("model key not found")
	}
	if time.Now().After(mk.ExpiresAt) {
		s.cache.DeleteModelKey(ctx, in.ModelApiKey)
		return nil, fmt.Errorf("model key has expired")
	}

	return &ValidateOutput{
		UserId:         mk.UserId,
		SessionId:      mk.SessionId,
		AgentType:      mk.AgentType,
		PluginInstance: mk.PluginInstance,
		ExpiresAt:      mk.ExpiresAt,
	}, nil
}

// RevokeModelKey revokes the model API key for a given session.
func (s *Service) RevokeModelKey(ctx context.Context, sessionId string) error {
	return s.cache.DeleteModelKeyBySession(ctx, sessionId)
}

// generateKey creates a random 32-char hex key with cmscoder_ prefix.
func generateKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "cmscoder_" + hex.EncodeToString(b)
}

// CompositeTokenFormat is the prefix for composite tokens.
const CompositeTokenFormat = "cmscoderv1_"

// GenerateCompositeToken creates a composite token binding modelApiKey and accessToken.
// Format: cmscoderv1_<base64(modelApiKey:accessToken)>
// This ensures both tokens must be present together to use the model endpoint.
func (s *Service) GenerateCompositeToken(sessionId, accessToken string) string {
	// Look up the model API key for this session.
	mk, err := s.cache.GetModelKeyBySession(context.Background(), sessionId)
	if err != nil || mk == nil {
		return ""
	}

	payload := mk.ModelApiKey + ":" + accessToken
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	return CompositeTokenFormat + encoded
}
