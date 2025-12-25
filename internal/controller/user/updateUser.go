package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/authz"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/errors"
)

// UpdateUserInfo updates a user by username.
// Permission rules:
//   - Admin: can update any user's all fields (username, nickname, avatar, email, password, status, role)
//   - Regular user: can only update their own basic fields (username, nickname, avatar, email)
//
// @Summary Update user
// @Description Update a user by username (partial update supported). Non-admin can only update self and only username/nickname/avatar/email; admin can update username/nickname/avatar/email/password/status/role.
// @Tags users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param user body v1.UpdateUserRequest true "User update payload"
// @Success 200 {object} v1.UserResponse "User updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users/{username} [put]
func (u *UserController) UpdateUserInfo(c *gin.Context) {
	klog.V(1).Info("user update function called.")

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

	// Load existing user first, so partial update won't wipe required fields
	existing, err := u.srv.Users().GetUserByUsername(context.Background(), username)
	if err != nil {
		klog.Errorf("failed to get user: %v", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Parse request body
	var req v1.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.Errorf("invalid request body: %v", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Start with existing user data
	user := *existing

	// --- Authorization (Casbin) ---
	// Determine scope (self/any) based on ownership.
	scope := authz.ScopeAny
	if requesterID != "" && requesterID == existing.ID {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceUser, scope)

	// Decide whether this request includes sensitive updates.
	hasSensitive := (req.Password != nil && *req.Password != "") || req.Status != nil || req.Role != nil

	// 1) Basic updates require user:update_basic
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionUserUpdateBasic)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// 2) Sensitive updates additionally require user:update_sensitive
	if hasSensitive {
		allowed, err := authz.Enforce(requesterRole, obj, authz.ActionUserUpdateSensitive)
		if err != nil {
			klog.Errorf("authz enforce failed: %v", err)
			core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
			return
		}
		if !allowed {
			core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
			return
		}
	}

	// Apply updates based on role permissions
	if requesterRole == model.UserRoleAdmin {
		// Admin can update all fields
		if req.Username != nil {
			user.Username = *req.Username
		}
		if req.Nickname != nil {
			user.Nickname = *req.Nickname
		}
		if req.Avatar != nil {
			user.Avatar = *req.Avatar
		}
		if req.Email != nil {
			user.Email = *req.Email
		}
		if req.Status != nil {
			user.Status = *req.Status
		}
		if req.Role != nil {
			user.Role = *req.Role
		}
		if req.Password != nil && *req.Password != "" {
			salt, err := passwd.GenerateSalt()
			if err != nil {
				klog.Errorf("failed to generate salt: %v", err)
				core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
				return
			}

			passwordHash, err := passwd.HashPassword(*req.Password, salt)
			if err != nil {
				klog.Errorf("failed to hash password: %v", err)
				core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
				return
			}

			user.Salt = salt
			user.PasswordHash = passwordHash
		}
	} else {
		// Regular user can only update basic fields (username, nickname, avatar, email)
		if req.Username != nil {
			user.Username = *req.Username
		}
		if req.Nickname != nil {
			user.Nickname = *req.Nickname
		}
		if req.Avatar != nil {
			user.Avatar = *req.Avatar
		}
		if req.Email != nil {
			user.Email = *req.Email
		}
		// Sensitive fields are already blocked by authz.ActionUserUpdateSensitive.
	}

	// Validate updated user data
	if errs := user.Validate(); len(errs) != 0 {
		klog.Errorf("validation failed: %v", errs)
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error()), nil)
		return
	}

	// Save updated user
	if err := u.srv.Users().UpdateUser(context.Background(), &user); err != nil {
		klog.Errorf("failed to update user: %v", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Return updated user info
	resp := v1.UserResponse{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	core.WriteResponse(c, nil, resp)
}
