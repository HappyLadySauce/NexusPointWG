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

// GetUserInfo get a user by id.
// @Summary Get user
// @Description Get a user by id
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} v1.UserResponse "User retrieved successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - missing id"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users/{id} [get]
func (u *UserController) GetUserInfo(c *gin.Context) {
	klog.V(1).Info("user get function called.")

	id := c.Param("id")
	if id == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing user id"), nil)
		return
	}

	user, err := u.srv.Users().GetUser(context.Background(), id)
	if err != nil {
		klog.Errorf("failed to get user: %v", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp := v1.UserResponse{
		Username: user.Username,
		Email:    user.Email,
	}

	core.WriteResponse(c, nil, resp)
}
