package user

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// ListUsers lists users with pagination and filters.
// @Summary List users
// @Description List users with optional filters (username, email, role, status) and pagination
// @Tags users
// @Produce json
// @Param username query string false "Filter by username (partial match)"
// @Param email query string false "Filter by email (partial match)"
// @Param role query string false "Filter by role (user/admin)"
// @Param status query string false "Filter by status (active/inactive/deleted)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 20, max: 200)"
// @Success 200 {object} v1.UserListResponse "Users listed successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid parameters"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users [get]
func (u *UserController) ListUsers(c *gin.Context) {
	klog.V(1).Info("user list function called.")

	opt := store.UserListOptions{
		Username: c.Query("username"),
		Email:    c.Query("email"),
		Role:     c.Query("role"),
		Status:   c.Query("status"),
	}

	// Parse pagination parameters
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			core.WriteResponse(c, errors.WithCode(code.ErrValidation, "invalid offset"), nil)
			return
		}
		opt.Offset = offset
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			core.WriteResponse(c, errors.WithCode(code.ErrValidation, "invalid limit"), nil)
			return
		}
		opt.Limit = limit
	}

	users, total, err := u.srv.Users().ListUsers(context.Background(), opt)
	if err != nil {
		klog.V(1).InfoS("failed to list users", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	items := make([]v1.UserResponse, 0, len(users))
	for _, user := range users {
		items = append(items, v1.UserResponse{
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
		})
	}

	resp := v1.UserListResponse{
		Total: total,
		Items: items,
	}

	core.WriteResponse(c, nil, resp)
}

