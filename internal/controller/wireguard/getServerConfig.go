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

// GetServerConfig gets the server WireGuard configuration (admin only).
// @Summary Get server configuration
// @Description Get the WireGuard server configuration. Admin only.
// @Tags wireguard
// @Produce json
// @Success 200 {object} v1.GetServerConfigResponse "Server configuration"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/server-config [get]
func (w *WGController) GetServerConfig(c *gin.Context) {
	klog.V(1).Info("wireguard server config get function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceWGServer, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGServerGet)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Get server config
	interfaceConfig, publicKey, serverIP, err := w.srv.WGServer().GetServerConfig(context.Background())
	if err != nil {
		klog.V(1).InfoS("failed to get server config", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Convert to response type
	resp := v1.GetServerConfigResponse{
		Address:    interfaceConfig.Address,
		ListenPort: interfaceConfig.ListenPort,
		PrivateKey: interfaceConfig.PrivateKey,
		MTU:        interfaceConfig.MTU,
		PostUp:     interfaceConfig.PostUp,
		PostDown:   interfaceConfig.PostDown,
		PublicKey:  publicKey,
		ServerIP:   serverIP,
	}

	klog.V(1).InfoS("wireguard server config retrieved successfully")
	core.WriteResponse(c, nil, resp)
}

