package wireguard

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// DeleteIPPool deletes an IP pool by ID (admin only).
// @Summary Delete IP pool
// @Description Delete an IP pool by ID. Admin only. The pool can only be deleted when no IPs are allocated from it.
// @Tags wireguard
// @Produce json
// @Param id path string true "IP Pool ID"
// @Success 200 {object} core.SuccessResponse "IP pool deleted successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - IP pool is in use"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - IP pool not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/ip-pools/{id} [delete]
func (w *WGController) DeleteIPPool(c *gin.Context) {
	klog.V(1).Info("wireguard ip pool delete function called.")

	poolID := c.Param("id")
	if poolID == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing IP pool ID"), nil)
		return
	}

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceIPPool, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionIPPoolDelete)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Delete IP pool (Service layer will check if pool is in use)
	if err := w.srv.IPPools().DeleteIPPool(context.Background(), poolID); err != nil {
		klog.V(1).InfoS("failed to delete IP pool", "poolID", poolID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("wireguard IP pool deleted successfully", "poolID", poolID)
	core.WriteResponse(c, nil, nil)
}
