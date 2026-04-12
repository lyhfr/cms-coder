package userprofile

import (
	"context"
	"time"

	"cmscoder-user-service/internal/cache"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"
)

// Service manages user profiles.
type Service struct {
	cache *cache.MemoryCache
}

// New creates a new user profile service.
func New(c *cache.MemoryCache) *Service {
	return &Service{cache: c}
}

// UpsertInput is the input for upserting a user.
type UpsertInput struct {
	IamUserId   string
	Email       string
	DisplayName string
	TenantId    string
}

// UpsertOutput is the output for upserting a user.
type UpsertOutput struct {
	UserId      string
	Email       string
	DisplayName string
	TenantId    string
}

// Upsert creates or updates a user profile.
func (s *Service) Upsert(ctx context.Context, in UpsertInput) (*UpsertOutput, error) {
	user, err := s.cache.GetUserByIamId(ctx, in.IamUserId)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to query user")
	}

	if user == nil {
		userId := guid.S()
		user = &cache.User{
			UserId:      userId,
			IamUserId:   in.IamUserId,
			Email:       in.Email,
			DisplayName: in.DisplayName,
			TenantId:    in.TenantId,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.cache.SetUser(ctx, user); err != nil {
			return nil, gerror.Wrap(err, "failed to create user")
		}
	} else {
		user.Email = in.Email
		user.DisplayName = in.DisplayName
		user.TenantId = in.TenantId
		user.UpdatedAt = time.Now()
		if err := s.cache.SetUser(ctx, user); err != nil {
			return nil, gerror.Wrap(err, "failed to update user")
		}
	}

	return &UpsertOutput{
		UserId:      user.UserId,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		TenantId:    user.TenantId,
	}, nil
}
