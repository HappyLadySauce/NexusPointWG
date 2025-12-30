package wireguard

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// GetPeer retrieves a WireGuard peer by ID.
// @Summary Get WireGuard peer
// @Description Get a WireGuard peer by ID. Admin can get any peer, regular users can only get their own peers.
// @Tags wireguard
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} v1.WGPeerResponse "Peer retrieved successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid peer ID"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - peer not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [get]
func (w *WGController) GetPeer(c *gin.Context) {
	klog.V(1).Info("wireguard peer get function called.")

	peerID := c.Param("id")
	if peerID == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing peer ID"), nil)
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

	// Get peer
	peer, err := w.srv.WGPeers().GetPeer(context.Background(), peerID)
	if err != nil {
		klog.V(1).InfoS("failed to get peer", "peerID", peerID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// --- Authorization (Casbin) ---
	scope := spec.ScopeAny
	if requesterID != "" && requesterID == peer.UserID {
		scope = spec.ScopeSelf
	}
	obj := spec.Obj(spec.ResourceWGPeer, scope)

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGPeerList)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Get user info for response
	user, _ := w.srv.Users().GetUser(context.Background(), peer.UserID)

	resp := v1.WGPeerResponse{
		ID:                  peer.ID,
		UserID:              peer.UserID,
		Username:            "",
		DeviceName:          peer.DeviceName,
		ClientPublicKey:     peer.ClientPublicKey,
		ClientIP:            peer.ClientIP,
		AllowedIPs:          peer.AllowedIPs,
		DNS:                 peer.DNS,
		Endpoint:            peer.Endpoint,
		PersistentKeepalive: peer.PersistentKeepalive,
		Status:              peer.Status,
		IPPoolID:            peer.IPPoolID,
		CreatedAt:           peer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           peer.UpdatedAt.Format(time.RFC3339),
	}
	if user != nil {
		resp.Username = user.Username
	}

	core.WriteResponse(c, nil, resp)
}
