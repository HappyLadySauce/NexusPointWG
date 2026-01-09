package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

type UserSrv interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, opt store.UserListOptions) ([]*model.User, int64, error)
	// BatchCreateUsers creates multiple users in a transaction.
	BatchCreateUsers(ctx context.Context, items []v1.CreateUserRequest, isAdmin bool) error
	// BatchUpdateUsers updates multiple users in a transaction.
	BatchUpdateUsers(ctx context.Context, items []v1.BatchUpdateUserItem) error
	// BatchDeleteUsers deletes multiple users by usernames in a transaction.
	BatchDeleteUsers(ctx context.Context, usernames []string) error
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

// BatchCreateUsers creates multiple users in a transaction.
// isAdmin indicates whether the requester is an admin (affects default role/status).
func (u *userSrv) BatchCreateUsers(ctx context.Context, items []v1.CreateUserRequest, isAdmin bool) error {
	users := make([]*model.User, 0, len(items))

	for _, item := range items {
		user := &model.User{}

		// Generate user ID
		userID, err := snowflake.GenerateID()
		if err != nil {
			klog.V(1).InfoS("failed to generate user ID", "username", item.Username, "error", err)
			return errors.WithCode(code.ErrUnknown, "failed to generate user ID")
		}
		user.ID = userID

		// Set username, nickname, avatar, email
		user.Username = item.Username
		if item.Nickname == "" {
			user.Nickname = item.Username
		} else {
			user.Nickname = item.Nickname
		}
		if item.Avatar == "" {
			user.Avatar = model.DefaultAvatarURL
		} else {
			user.Avatar = item.Avatar
		}
		user.Email = item.Email

		// Set role and status based on isAdmin flag
		if isAdmin {
			if item.Role != nil {
				user.Role = *item.Role
			} else {
				user.Role = model.UserRoleUser
			}
			if item.Status != nil {
				user.Status = *item.Status
			} else {
				user.Status = model.UserStatusActive
			}
		} else {
			// Public registration: only regular users
			user.Role = model.UserRoleUser
			user.Status = model.UserStatusActive
		}

		// Hash password
		if item.Password != "" {
			salt, err := passwd.GenerateSalt()
			if err != nil {
				klog.V(1).InfoS("failed to generate salt", "username", item.Username, "error", err)
				return errors.WithCode(code.ErrEncrypt, "%s", err.Error())
			}

			passwordHash, err := passwd.HashPassword(item.Password, salt)
			if err != nil {
				klog.V(1).InfoS("failed to hash password", "username", item.Username, "error", err)
				return errors.WithCode(code.ErrEncrypt, "%s", err.Error())
			}

			user.Salt = salt
			user.PasswordHash = passwordHash
		}

		// Ensure defaults
		ensureUserDefaults(user)

		// Validate user
		if errs := user.Validate(); len(errs) != 0 {
			klog.V(1).InfoS("validation failed", "username", item.Username, "errors", errs.ToAggregate().Error())
			return errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error())
		}

		users = append(users, user)
	}

	// Batch create in transaction
	return u.store.Users().BatchCreateUsers(ctx, users)
}

// BatchUpdateUsers updates multiple users in a transaction.
func (u *userSrv) BatchUpdateUsers(ctx context.Context, items []v1.BatchUpdateUserItem) error {
	users := make([]*model.User, 0, len(items))

	for _, item := range items {
		// Get existing user
		existing, err := u.store.Users().GetUserByUsername(ctx, item.Username)
		if err != nil {
			return err
		}

		// Apply updates
		// Note: item.Username is the identifier (string), not the field to update
		// The UpdateUserRequest fields are embedded, so we access them directly
		// To update username, we would need a different field name, but for now
		// username updates are not supported in batch operations to avoid confusion
		if item.Nickname != nil {
			existing.Nickname = *item.Nickname
		}
		if item.Avatar != nil {
			existing.Avatar = *item.Avatar
		}
		if item.Email != nil {
			existing.Email = *item.Email
		}
		if item.Password != nil && *item.Password != "" {
			salt, err := passwd.GenerateSalt()
			if err != nil {
				klog.V(1).InfoS("failed to generate salt", "username", item.Username, "error", err)
				return errors.WithCode(code.ErrEncrypt, "%s", err.Error())
			}

			passwordHash, err := passwd.HashPassword(*item.Password, salt)
			if err != nil {
				klog.V(1).InfoS("failed to hash password", "username", item.Username, "error", err)
				return errors.WithCode(code.ErrEncrypt, "%s", err.Error())
			}

			existing.Salt = salt
			existing.PasswordHash = passwordHash
		}
		if item.Status != nil {
			existing.Status = *item.Status
		}
		if item.Role != nil {
			existing.Role = *item.Role
		}

		// Ensure defaults
		ensureUserDefaults(existing)

		users = append(users, existing)
	}

	// Batch update in transaction
	return u.store.Users().BatchUpdateUsers(ctx, users)
}

// BatchDeleteUsers deletes multiple users by usernames in a transaction.
func (u *userSrv) BatchDeleteUsers(ctx context.Context, usernames []string) error {
	ids := make([]string, 0, len(usernames))

	// Convert usernames to IDs
	for _, username := range usernames {
		user, err := u.store.Users().GetUserByUsername(ctx, username)
		if err != nil {
			return err
		}
		ids = append(ids, user.ID)
	}

	// Batch delete in transaction
	return u.store.Users().BatchDeleteUsers(ctx, ids)
}
