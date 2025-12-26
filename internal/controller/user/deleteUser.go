package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// DeleteUser deletes a user by username.
// Permission rules:
//   - Admin: can hard delete any user (permanent removal from database)
//   - Regular user: can only soft delete themselves (set status=deleted)
//
// @Summary Delete user
// @Description Delete a user by username. Non-admin can only logout (soft delete) self by setting status=deleted; admin can hard delete any user.
// @Tags users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} core.SuccessResponse "User deleted successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - missing username"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users/{username} [delete]
func (u *UserController) DeleteUser(c *gin.Context) {
	klog.V(1).Info("user delete function called.")

	// Validate username parameter
	username := c.Param("username")
	if username == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing username"), nil)
		return
	}

	// Get requester info from JWTAuth middleware
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	// Get target user by username
	targetUser, err := u.srv.Users().GetUserByUsername(context.Background(), username)
	if err != nil {
		klog.V(1).InfoS("failed to get user", "username", username, "requesterID", requesterID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// --- Authorization (Casbin) ---
	scope := spec.ScopeAny
	if requesterID != "" && requesterID == targetUser.ID {
		scope = spec.ScopeSelf
	}
	obj := spec.Obj(spec.ResourceUser, scope)

	// Admin: hard delete any. Regular: soft delete self.
	act := spec.ActionUserSoftDelete
	if requesterRole == model.UserRoleAdmin {
		act = spec.ActionUserHardDelete
	}
	allowed, err := spec.Enforce(requesterRole, obj, act)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "username", username, "requesterID", requesterID, "requesterRole", requesterRole, "action", act, "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Execute delete operation based on role
	if requesterRole == model.UserRoleAdmin {
		// Admin: hard delete (permanent removal from database)
		if err := u.srv.Users().DeleteUser(context.Background(), targetUser.ID); err != nil {
			klog.V(1).InfoS("failed to hard delete user", "username", username, "userID", targetUser.ID, "requesterID", requesterID, "error", err)
			core.WriteResponse(c, err, nil)
			return
		}
		klog.V(1).InfoS("admin hard deleted user", "username", username, "userID", targetUser.ID, "requesterID", requesterID)
	} else {
		// Regular user: soft delete (set status=deleted)
		targetUser.Status = model.UserStatusDeleted
		if err := u.srv.Users().UpdateUser(context.Background(), targetUser); err != nil {
			klog.V(1).InfoS("failed to soft delete user", "username", username, "userID", targetUser.ID, "requesterID", requesterID, "error", err)
			core.WriteResponse(c, err, nil)
			return
		}
		klog.V(1).InfoS("user soft deleted themselves", "username", username, "userID", targetUser.ID)
	}

	core.WriteResponse(c, nil, nil)
}
