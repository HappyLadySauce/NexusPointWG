package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/errors"
)

// ChangePassword changes a user's password.
// Permission rules:
//   - Users can only change their own password (self-only)
//   - Old password must be verified before allowing the change
//
// @Summary Change password
// @Description Change user password by username. User must provide old password for verification and can only change their own password.
// @Tags users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param request body v1.ChangePwdRequest true "Change password request"
// @Success 200 {object} core.SuccessResponse "Password changed successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token, or incorrect old password"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error or encryption error"
// @Router /api/v1/users/{username}/password [post]
func (u *UserController) ChangePassword(c *gin.Context) {
	klog.V(1).Info("user change password function called.")

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

	// Parse request body
	var req v1.ChangePwdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "username", username, "requesterID", requesterID, "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// --- Authorization (Casbin) ---
	// Users can only change their own password (self-only)
	scope := spec.ScopeAny
	if requesterID != "" && requesterID == targetUser.ID {
		scope = spec.ScopeSelf
	}
	obj := spec.Obj(spec.ResourceUser, scope)

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserChangePassword)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "username", username, "requesterID", requesterID, "requesterRole", requesterRole, "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Verify old password
	if !passwd.VerifyPassword(req.OldPassword, targetUser.Salt, targetUser.PasswordHash) {
		klog.V(1).InfoS("incorrect old password", "username", username, "requesterID", requesterID)
		core.WriteResponse(c, errors.WithCode(code.ErrPasswordIncorrect, "%s", code.Message(code.ErrPasswordIncorrect)), nil)
		return
	}

	// Generate new salt and hash new password
	salt, err := passwd.GenerateSalt()
	if err != nil {
		klog.V(1).InfoS("failed to generate salt", "username", username, "requesterID", requesterID, "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
		return
	}

	passwordHash, err := passwd.HashPassword(req.NewPassword, salt)
	if err != nil {
		klog.V(1).InfoS("failed to hash password", "username", username, "requesterID", requesterID, "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
		return
	}

	// Update user with new password hash and salt
	targetUser.Salt = salt
	targetUser.PasswordHash = passwordHash

	// Save updated user
	if err := u.srv.Users().UpdateUser(context.Background(), targetUser); err != nil {
		klog.V(1).InfoS("failed to update password", "username", username, "userID", targetUser.ID, "requesterID", requesterID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("password changed successfully", "username", username, "userID", targetUser.ID, "requesterID", requesterID)
	core.WriteResponse(c, nil, nil)
}
