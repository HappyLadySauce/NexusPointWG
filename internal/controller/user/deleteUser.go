package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// DeleteUser deletes a user by username.
// @Summary Delete user
// @Description Delete a user by username. Non-admin can only logout (soft delete) self by setting status=deleted; admin can hard delete any user.
// @Tags users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} nil "User deleted successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - missing username"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users/{username} [delete]
func (u *UserController) DeleteUser(c *gin.Context) {
	klog.V(1).Info("user delete function called.")

	username := c.Param("username")
	if username == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing username"), nil)
		return
	}

	// Get requester info from JWTAuth middleware.
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)

	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	// Get user by username first
	user, err := u.srv.Users().GetUserByUsername(context.Background(), username)
	if err != nil {
		klog.Errorf("failed to get user: %s", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Admin can hard-delete any user.
	if requesterRole == model.UserRoleAdmin {
		if err := u.srv.Users().DeleteUser(context.Background(), user.ID); err != nil {
			klog.Errorf("failed to hard delete user: %s", err)
			core.WriteResponse(c, err, nil)
			return
		}
		core.WriteResponse(c, nil, nil)
		return
	}

	// Non-admin can only soft-delete (logout) themselves.
	if requesterID == "" || user.ID != requesterID {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}
	user.Status = model.UserStatusDeleted

	if err := u.srv.Users().UpdateUser(context.Background(), user); err != nil {
		klog.Errorf("failed to soft delete user: %s", err)
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, nil)
}
