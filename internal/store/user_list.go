package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

type UserListOptions struct {
	Username string
	Email    string
	Role     string
	Status   string
	Offset   int
	Limit    int
}

type UserListItem struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Status   string `json:"status"`
	Role     string `json:"role"`
	Avatar   string `json:"avatar"`
}

type UserListResponse struct {
	Total int64          `json:"total"`
	Items []UserListItem `json:"items"`
}

// ListUsers returns users by keyword (username/email/nickname) with pagination.
type UserLister interface {
	ListUsers(ctx context.Context, opt UserListOptions) ([]*model.User, int64, error)
}
