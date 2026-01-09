package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

type UserStore interface {
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, opt UserListOptions) ([]*model.User, int64, error)
	// BatchCreateUsers creates multiple users in a transaction. Returns error if any user creation fails.
	BatchCreateUsers(ctx context.Context, users []*model.User) error
	// BatchUpdateUsers updates multiple users in a transaction. Returns error if any user update fails.
	BatchUpdateUsers(ctx context.Context, users []*model.User) error
	// BatchDeleteUsers deletes multiple users by IDs in a transaction. Returns error if any user deletion fails.
	BatchDeleteUsers(ctx context.Context, ids []string) error
}

type UserListOptions struct {
	Username string
	Email    string
	Role     string
	Status   string
	Offset   int
	Limit    int
}
