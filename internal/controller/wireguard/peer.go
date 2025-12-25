package wireguard

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/authz"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// ListPeers lists peers.
// - admin: list all peers
// - user: list own peers
func (w *WireGuardController) ListPeers(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	scope := authz.ScopeSelf
	if requesterRole == model.UserRoleAdmin {
		scope = authz.ScopeAny
	}
	obj := authz.Obj(authz.ResourceWGPeer, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGPeerList)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	var peers []*model.WGPeer
	if requesterRole == model.UserRoleAdmin {
		peers, err = w.srv.WGPeers().ListPeers(context.Background())
	} else {
		peers, err = w.srv.WGPeers().ListPeersByUser(context.Background(), requesterID)
	}
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	resp := make([]v1.WGPeerResponse, 0, len(peers))
	for _, p := range peers {
		resp = append(resp, v1.WGPeerResponse{
			ID:              p.ID,
			UserID:          p.UserID,
			Name:            p.Name,
			AllowedIPs:      p.AllowedIPs,
			ClientPublicKey: p.ClientPublicKey,
		})
	}
	core.WriteResponse(c, nil, resp)
}

// CreatePeer creates a new peer.
// - admin: can create for any user (by user_id)
// - user: can only create for self (user_id ignored)
func (w *WireGuardController) CreatePeer(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	var req v1.CreateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	ownerID := requesterID
	if requesterRole == model.UserRoleAdmin && req.UserID != "" {
		ownerID = req.UserID
	}

	scope := authz.ScopeAny
	if ownerID == requesterID && requesterID != "" {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceWGPeer, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGPeerCreate)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	node, _ := snowflake.NewNode(1)
	peer := &model.WGPeer{
		ID:              node.Generate().String(),
		UserID:          ownerID,
		Name:            req.Name,
		AllowedIPs:      req.AllowedIPs,
		ClientPublicKey: req.ClientPublicKey,
	}
	if errs := peer.Validate(); len(errs) != 0 {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error()), nil)
		return
	}
	if err := w.srv.WGPeers().CreatePeer(context.Background(), peer); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, v1.WGPeerResponse{
		ID:              peer.ID,
		UserID:          peer.UserID,
		Name:            peer.Name,
		AllowedIPs:      peer.AllowedIPs,
		ClientPublicKey: peer.ClientPublicKey,
	})
}

// GetPeer gets a peer by id.
func (w *WireGuardController) GetPeer(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	id := c.Param("id")
	if id == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing peer id"), nil)
		return
	}

	peer, err := w.srv.WGPeers().GetPeer(context.Background(), id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	scope := authz.ScopeAny
	if requesterID != "" && requesterID == peer.UserID {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceWGPeer, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGPeerRead)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	core.WriteResponse(c, nil, v1.WGPeerResponse{
		ID:              peer.ID,
		UserID:          peer.UserID,
		Name:            peer.Name,
		AllowedIPs:      peer.AllowedIPs,
		ClientPublicKey: peer.ClientPublicKey,
	})
}

// UpdatePeer updates a peer by id (name/allowedIPs/publicKey).
func (w *WireGuardController) UpdatePeer(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	id := c.Param("id")
	if id == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing peer id"), nil)
		return
	}

	peer, err := w.srv.WGPeers().GetPeer(context.Background(), id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	scope := authz.ScopeAny
	if requesterID != "" && requesterID == peer.UserID {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceWGPeer, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGPeerUpdate)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	var req v1.UpdateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponseBindErr(c, err, nil)
		return
	}
	if req.Name != nil {
		peer.Name = *req.Name
	}
	if req.AllowedIPs != nil {
		peer.AllowedIPs = *req.AllowedIPs
	}
	if req.ClientPublicKey != nil {
		peer.ClientPublicKey = *req.ClientPublicKey
	}

	if errs := peer.Validate(); len(errs) != 0 {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error()), nil)
		return
	}
	if err := w.srv.WGPeers().UpdatePeer(context.Background(), peer); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, v1.WGPeerResponse{
		ID:              peer.ID,
		UserID:          peer.UserID,
		Name:            peer.Name,
		AllowedIPs:      peer.AllowedIPs,
		ClientPublicKey: peer.ClientPublicKey,
	})
}

// DeletePeer deletes a peer by id.
func (w *WireGuardController) DeletePeer(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	id := c.Param("id")
	if id == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing peer id"), nil)
		return
	}

	peer, err := w.srv.WGPeers().GetPeer(context.Background(), id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	scope := authz.ScopeAny
	if requesterID != "" && requesterID == peer.UserID {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceWGPeer, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGPeerDelete)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	if err := w.srv.WGPeers().DeletePeer(context.Background(), id); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}

// DownloadPeerConfig returns a minimal config payload for a peer.
// NOTE: This is a scaffold. Real WireGuard config generation should be implemented in a dedicated package.
func (w *WireGuardController) DownloadPeerConfig(c *gin.Context) {
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	id := c.Param("id")
	if id == "" {
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "missing peer id"), nil)
		return
	}
	peer, err := w.srv.WGPeers().GetPeer(context.Background(), id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	scope := authz.ScopeAny
	if requesterID != "" && requesterID == peer.UserID {
		scope = authz.ScopeSelf
	}
	obj := authz.Obj(authz.ResourceWGConfig, scope)
	allowed, err := authz.Enforce(requesterRole, obj, authz.ActionWGConfigDownload)
	if err != nil {
		klog.Errorf("authz enforce failed: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Placeholder config.
	cfg := "[Interface]\n" +
		"# TODO: generate real client private key\n" +
		"Address = " + peer.AllowedIPs + "\n\n" +
		"[Peer]\n" +
		"# TODO: set real server public key and endpoint\n" +
		"AllowedIPs = 0.0.0.0/0, ::/0\n"

	core.WriteResponse(c, nil, v1.WGConfigResponse{PeerID: peer.ID, Config: cfg})
}
