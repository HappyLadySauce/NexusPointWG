package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

type UserSrv interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, opt store.UserListOptions) ([]*model.User, int64, error)
}

type userSrv struct {
	store store.Factory
}

// UserSrv if implemented, then userSrv implements UserSrv interface.
var _ UserSrv = (*userSrv)(nil)

func newUsers(s *service) *userSrv {
	return &userSrv{store: s.store}
}

func ensureUserDefaults(user *model.User) {
	if user == nil {
		return
	}
	// If role is still empty (e.g. legacy records), default to normal user.
	if user.Role == "" {
		user.Role = model.UserRoleUser
	}
	// Keep behavior consistent with existing admin update logic.
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
}

func (u *userSrv) CreateUser(ctx context.Context, user *model.User) error {
	// Database layer already handles unique constraint errors and returns appropriate error codes
	ensureUserDefaults(user)
	return u.store.Users().CreateUser(ctx, user)
}

func (u *userSrv) GetUser(ctx context.Context, id string) (*model.User, error) {
	return u.store.Users().GetUser(ctx, id)
}

func (u *userSrv) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return u.store.Users().GetUserByUsername(ctx, username)
}

func (u *userSrv) UpdateUser(ctx context.Context, user *model.User) error {
	// Database layer already handles errors and returns appropriate error codes
	ensureUserDefaults(user)
	return u.store.Users().UpdateUser(ctx, user)
}

func (u *userSrv) DeleteUser(ctx context.Context, id string) error {
	return u.store.Users().DeleteUser(ctx, id)
}

func (u *userSrv) ListUsers(ctx context.Context, opt store.UserListOptions) ([]*model.User, int64, error) {
	return u.store.Users().ListUsers(ctx, opt)
}
