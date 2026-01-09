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
	"github.com/HappyLadySauce/errors"
)

const maxBatchSize = 50

// BatchCreateUsers creates multiple users in a transaction (admin only).
// @Summary Batch create users
// @Description Create multiple users in a single transaction. Admin only. Maximum 50 users per request.
// @Tags users
// @Accept json
// @Produce json
// @Param users body v1.BatchCreateUsersRequest true "List of users to create"
// @Success 200 {object} v1.BatchCreateUsersResponse "Users created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/users/batch [post]
func (u *UserController) BatchCreateUsers(c *gin.Context) {
	klog.V(1).Info("batch user create function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceUser, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserCreate)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		klog.V(1).InfoS("permission denied for batch user creation", "requesterRole", requesterRole)
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Parse request body
	var req v1.BatchCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Validate batch size
	if len(req.Items) == 0 {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "items list cannot be empty"), nil)
		return
	}
	if len(req.Items) > maxBatchSize {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "batch size exceeds maximum of %d", maxBatchSize), nil)
		return
	}

	// Check if requester is admin
	isAdmin := requesterRole == "admin"

	// Call Service layer to batch create users
	if err := u.srv.Users().BatchCreateUsers(context.Background(), req.Items, isAdmin); err != nil {
		klog.V(1).InfoS("failed to batch create users", "count", len(req.Items), "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("batch users created successfully", "count", len(req.Items), "requesterRole", requesterRole)
	resp := v1.BatchCreateUsersResponse{
		Count: int64(len(req.Items)),
	}
	core.WriteResponse(c, nil, resp)
}

// BatchUpdateUsers updates multiple users in a transaction.
// @Summary Batch update users
// @Description Update multiple users in a single transaction. Maximum 50 users per request.
// @Tags users
// @Accept json
// @Produce json
// @Param users body v1.BatchUpdateUsersRequest true "List of users to update"
// @Success 200 {object} v1.BatchUpdateUsersResponse "Users updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/users/batch [put]
func (u *UserController) BatchUpdateUsers(c *gin.Context) {
	klog.V(1).Info("batch user update function called.")

	// Get requester info from JWTAuth middleware
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	// Parse request body
	var req v1.BatchUpdateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Validate batch size
	if len(req.Items) == 0 {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "items list cannot be empty"), nil)
		return
	}
	if len(req.Items) > maxBatchSize {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "batch size exceeds maximum of %d", maxBatchSize), nil)
		return
	}

	// Check permissions for each user update
	for _, item := range req.Items {
		// Load existing user
		existing, err := u.srv.Users().GetUserByUsername(context.Background(), item.Username)
		if err != nil {
			klog.V(1).InfoS("failed to get user for permission check", "username", item.Username, "error", err)
			core.WriteResponse(c, err, nil)
			return
		}

		// Determine scope
		scope := spec.ScopeAny
		if requesterID != "" && requesterID == existing.ID {
			scope = spec.ScopeSelf
		}
		obj := spec.Obj(spec.ResourceUser, scope)

		// Check if request includes sensitive updates
		hasSensitive := (item.Password != nil && *item.Password != "") || item.Status != nil || item.Role != nil

		// 1) Basic updates require user:update_basic
		allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserUpdateBasic)
		if err != nil {
			klog.V(1).InfoS("authz enforce failed", "username", item.Username, "error", err)
			core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
			return
		}
		if !allowed {
			klog.V(1).InfoS("permission denied for user update", "username", item.Username, "requesterRole", requesterRole)
			core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
			return
		}

		// 2) Sensitive updates additionally require user:update_sensitive
		if hasSensitive {
			allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserUpdateSensitive)
			if err != nil {
				klog.V(1).InfoS("authz enforce failed", "username", item.Username, "error", err)
				core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
				return
			}
			if !allowed {
				klog.V(1).InfoS("permission denied for sensitive user update", "username", item.Username, "requesterRole", requesterRole)
				core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
				return
			}
		}
	}

	// Call Service layer to batch update users
	if err := u.srv.Users().BatchUpdateUsers(context.Background(), req.Items); err != nil {
		klog.V(1).InfoS("failed to batch update users", "count", len(req.Items), "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("batch users updated successfully", "count", len(req.Items), "requesterID", requesterID)
	resp := v1.BatchUpdateUsersResponse{
		Count: int64(len(req.Items)),
	}
	core.WriteResponse(c, nil, resp)
}

// BatchDeleteUsers deletes multiple users in a transaction (admin only).
// @Summary Batch delete users
// @Description Delete multiple users in a single transaction. Admin only. Maximum 50 users per request.
// @Tags users
// @Accept json
// @Produce json
// @Param users body v1.BatchDeleteUsersRequest true "List of usernames to delete"
// @Success 200 {object} v1.BatchDeleteUsersResponse "Users deleted successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/users/batch [delete]
func (u *UserController) BatchDeleteUsers(c *gin.Context) {
	klog.V(1).Info("batch user delete function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceUser, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserHardDelete)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		klog.V(1).InfoS("permission denied for batch user deletion", "requesterRole", requesterRole)
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Parse request body
	var req v1.BatchDeleteUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Validate batch size
	if len(req.Usernames) == 0 {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "usernames list cannot be empty"), nil)
		return
	}
	if len(req.Usernames) > maxBatchSize {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "batch size exceeds maximum of %d", maxBatchSize), nil)
		return
	}

	// Call Service layer to batch delete users
	if err := u.srv.Users().BatchDeleteUsers(context.Background(), req.Usernames); err != nil {
		klog.V(1).InfoS("failed to batch delete users", "count", len(req.Usernames), "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("batch users deleted successfully", "count", len(req.Usernames), "requesterRole", requesterRole)
	resp := v1.BatchDeleteUsersResponse{
		Count: int64(len(req.Usernames)),
	}
	core.WriteResponse(c, nil, resp)
}
