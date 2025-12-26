package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// GetUserInfo get a user by username.
// @Summary Get user
// @Description Get a user by username
// @Tags users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} v1.UserResponse "User retrieved successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - missing username"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users/{username} [get]
func (u *UserController) GetUserInfo(c *gin.Context) {
	klog.V(1).Info("user get function called.")

	username := c.Param("username")
	if username == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing username"), nil)
		return
	}

	user, err := u.srv.Users().GetUserByUsername(context.Background(), username)
	if err != nil {
		klog.V(1).InfoS("failed to get user", "username", username, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp := v1.UserResponse{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	core.WriteResponse(c, nil, resp)
}
