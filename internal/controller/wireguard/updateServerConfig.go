package wireguard

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// UpdateServerConfig updates the server WireGuard configuration (admin only).
// @Summary Update server configuration
// @Description Update the WireGuard server configuration. Admin only. Updates will automatically sync to all client configurations.
// @Tags wireguard
// @Accept json
// @Produce json
// @Param request body v1.UpdateServerConfigRequest true "Server configuration update request"
// @Success 200 {object} core.SuccessResponse "Server configuration updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/server-config [put]
func (w *WGController) UpdateServerConfig(c *gin.Context) {
	klog.V(1).Info("wireguard server config update function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceWGServer, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGServerUpdate)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Bind request
	var req v1.UpdateServerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Update server config
	if err := w.srv.WGServer().UpdateServerConfig(context.Background(), &req); err != nil {
		klog.V(1).InfoS("failed to update server config", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("wireguard server config updated successfully")
	core.WriteResponse(c, nil, nil)
}

