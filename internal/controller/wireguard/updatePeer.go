package wireguard

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// UpdatePeer updates a WireGuard peer by ID.
// @Summary Update WireGuard peer
// @Description Update a WireGuard peer by ID. Admin can update any peer, regular users can only update their own peers.
// @Tags wireguard
// @Accept json
// @Produce json
// @Param id path string true "Peer ID"
// @Param peer body v1.UpdateWGPeerRequest true "Peer update information"
// @Success 200 {object} v1.WGPeerResponse "Peer updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - peer not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [put]
func (w *WGController) UpdatePeer(c *gin.Context) {
	klog.V(1).Info("wireguard peer update function called.")

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

	// Get existing peer
	existingPeer, err := w.srv.WGPeers().GetPeer(context.Background(), peerID)
	if err != nil {
		klog.V(1).InfoS("failed to get peer", "peerID", peerID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// --- Authorization (Casbin) ---
	scope := spec.ScopeAny
	if requesterID != "" && requesterID == existingPeer.UserID {
		scope = spec.ScopeSelf
	}
	obj := spec.Obj(spec.ResourceWGPeer, scope)

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGPeerUpdate)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Parse request body
	var req v1.UpdateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Update peer fields (only update provided fields)
	if req.DeviceName != nil {
		existingPeer.DeviceName = *req.DeviceName
	}
	if req.ClientIP != nil {
		existingPeer.ClientIP = *req.ClientIP
	}
	if req.IPPoolID != nil {
		existingPeer.IPPoolID = *req.IPPoolID
	}
	if req.ClientPrivateKey != nil {
		existingPeer.ClientPrivateKey = *req.ClientPrivateKey
		// Regenerate public key from private key
		publicKey, err := wireguard.GeneratePublicKey(*req.ClientPrivateKey)
		if err != nil {
			klog.V(1).InfoS("failed to generate public key from private key", "error", err)
			core.WriteResponse(c, errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "failed to generate public key from private key"), nil)
			return
		}
		existingPeer.ClientPublicKey = publicKey
	}
	if req.AllowedIPs != nil {
		existingPeer.AllowedIPs = *req.AllowedIPs
	}
	if req.DNS != nil {
		existingPeer.DNS = *req.DNS
	}
	if req.Endpoint != nil {
		existingPeer.Endpoint = *req.Endpoint
	}
	if req.PersistentKeepalive != nil {
		existingPeer.PersistentKeepalive = *req.PersistentKeepalive
	}
	if req.Status != nil {
		// Validate status
		if *req.Status != model.WGPeerStatusActive && *req.Status != model.WGPeerStatusDisabled {
			core.WriteResponse(c, errors.WithCode(code.ErrValidation, "invalid status, must be 'active' or 'disabled'"), nil)
			return
		}
		existingPeer.Status = *req.Status
	}

	// Update peer (service layer handles IP allocation and key validation)
	if err := w.srv.WGPeers().UpdatePeer(context.Background(), existingPeer, req.ClientIP, req.IPPoolID); err != nil {
		klog.V(1).InfoS("failed to update peer", "peerID", peerID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Get updated peer for response
	updatedPeer, err := w.srv.WGPeers().GetPeer(context.Background(), peerID)
	if err != nil {
		klog.V(1).InfoS("failed to get updated peer", "peerID", peerID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Get user info for response
	user, _ := w.srv.Users().GetUser(context.Background(), updatedPeer.UserID)

	resp := v1.WGPeerResponse{
		ID:                  updatedPeer.ID,
		UserID:              updatedPeer.UserID,
		Username:            "",
		DeviceName:          updatedPeer.DeviceName,
		ClientPublicKey:     updatedPeer.ClientPublicKey,
		ClientIP:            updatedPeer.ClientIP,
		AllowedIPs:          updatedPeer.AllowedIPs,
		DNS:                 updatedPeer.DNS,
		Endpoint:            updatedPeer.Endpoint,
		PersistentKeepalive: updatedPeer.PersistentKeepalive,
		Status:              updatedPeer.Status,
		IPPoolID:            updatedPeer.IPPoolID,
		CreatedAt:           updatedPeer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           updatedPeer.UpdatedAt.Format(time.RFC3339),
	}
	if user != nil {
		resp.Username = user.Username
	}

	klog.V(1).InfoS("wireguard peer updated successfully", "peerID", peerID)
	core.WriteResponse(c, nil, resp)
}
