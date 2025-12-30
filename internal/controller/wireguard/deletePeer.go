package wireguard

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

// DeletePeer deletes a WireGuard peer by ID.
// @Summary Delete WireGuard peer
// @Description Delete a WireGuard peer by ID. Admin can delete any peer, regular users can only delete their own peers.
// @Tags wireguard
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} core.SuccessResponse "Peer deleted successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid peer ID"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - peer not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [delete]
func (w *WGController) DeletePeer(c *gin.Context) {
	klog.V(1).Info("wireguard peer delete function called.")

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

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGPeerDelete)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Determine if this is a hard delete (admin only)
	isHardDelete := requesterRole == model.UserRoleAdmin

	// Delete peer (IP allocation release/delete is handled in Service layer)
	if err := w.srv.WGPeers().DeletePeer(context.Background(), peerID, isHardDelete); err != nil {
		klog.V(1).InfoS("failed to delete peer", "peerID", peerID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).InfoS("wireguard peer deleted successfully", "peerID", peerID, "hardDelete", isHardDelete)
	core.WriteResponse(c, nil, nil)
}
