package wireguard

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/errors"
)

// CreatePeer creates a new WireGuard peer.
// @Summary Create WireGuard peer
// @Description Create a new WireGuard peer for a user. Admin can create peers for any user, regular users can only create peers for themselves.
// @Tags wireguard
// @Accept json
// @Produce json
// @Param peer body v1.CreateWGPeerRequest true "Peer information"
// @Success 200 {object} v1.WGPeerResponse "Peer created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers [post]
func (w *WGController) CreatePeer(c *gin.Context) {
	klog.V(1).Info("wireguard peer create function called.")

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
	var req v1.CreateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Determine target user ID
	targetUserID := requesterID
	if requesterRole == model.UserRoleAdmin {
		if req.Username != "" {
			// Look up user by username
			user, err := w.srv.Users().GetUserByUsername(context.Background(), req.Username)
			if err != nil {
				core.WriteResponse(c, errors.WithCode(code.ErrUserNotFound, "user not found: %s", req.Username), nil)
				return
			}
			targetUserID = user.ID
		} else if req.UserID != "" {
			// Backward compatibility: support UserID
		targetUserID = req.UserID
		}
	}

	// --- Authorization (Casbin) ---
	scope := spec.ScopeSelf
	if requesterRole == model.UserRoleAdmin && targetUserID != requesterID {
		scope = spec.ScopeAny
	}
	obj := spec.Obj(spec.ResourceWGPeer, scope)

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGPeerCreate)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Call Service layer to create peer (includes IP allocation and key generation)
	peer, err := w.srv.WGPeers().CreatePeer(
		context.Background(),
		targetUserID,
		req.DeviceName,
		req.IPPoolID,
		req.ClientIP,
		req.AllowedIPs,
		req.DNS,
		req.Endpoint,
		req.ClientPrivateKey,
		req.PersistentKeepalive,
	)
	if err != nil {
		klog.V(1).InfoS("failed to create peer", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// TODO: Generate client config file
	// TODO: Update server config file
	// TODO: Apply server config

	// Get user info for response
	user, _ := w.srv.Users().GetUser(context.Background(), targetUserID)

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

	// TODO: Generate client config file
	// TODO: Update server config file
	// TODO: Apply server config

	klog.V(1).InfoS("wireguard peer created successfully", "peerID", peer.ID, "userID", targetUserID)
	core.WriteResponse(c, nil, resp)
}

// ListPeers lists WireGuard peers with pagination and filters.
// @Summary List WireGuard peers
// @Description List WireGuard peers with optional filters and pagination. Admin can see all peers, regular users can only see their own peers.
// @Tags wireguard
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param status query string false "Filter by status (active/disabled)"
// @Param ip_pool_id query string false "Filter by IP pool ID"
// @Param device_name query string false "Filter by device name (partial match)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 20, max: 200)"
// @Success 200 {object} v1.WGPeerListResponse "Peers listed successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid parameters"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers [get]
func (w *WGController) ListPeers(c *gin.Context) {
	klog.V(1).Info("wireguard peer list function called.")

	// Get requester info from JWTAuth middleware
	requesterIDAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "missing auth context"), nil)
		return
	}
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterID, _ := requesterIDAny.(string)
	requesterRole, _ := requesterRoleAny.(string)

	// Build list options
	opt := store.WGPeerListOptions{
		UserID:     c.Query("user_id"),
		Status:     c.Query("status"),
		IPPoolID:   c.Query("ip_pool_id"),
		DeviceName: c.Query("device_name"),
	}

	// Regular users can only see their own peers
	if requesterRole != model.UserRoleAdmin {
		opt.UserID = requesterID
	}

	// Parse pagination parameters
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			core.WriteResponse(c, errors.WithCode(code.ErrValidation, "invalid offset"), nil)
			return
		}
		opt.Offset = offset
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			core.WriteResponse(c, errors.WithCode(code.ErrValidation, "invalid limit"), nil)
			return
		}
		opt.Limit = limit
	}

	// --- Authorization (Casbin) ---
	scope := spec.ScopeSelf
	if requesterRole == model.UserRoleAdmin {
		scope = spec.ScopeAny
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

	// List peers
	peers, total, err := w.srv.WGPeers().ListPeers(context.Background(), opt)
	if err != nil {
		klog.V(1).InfoS("failed to list peers", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Convert to response format
	items := make([]v1.WGPeerResponse, 0, len(peers))
	for _, peer := range peers {
		// Get user info
		user, _ := w.srv.Users().GetUser(context.Background(), peer.UserID)

		items = append(items, v1.WGPeerResponse{
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
		})
		if user != nil {
			items[len(items)-1].Username = user.Username
		}
	}

	resp := v1.WGPeerListResponse{
		Total: total,
		Items: items,
	}

	core.WriteResponse(c, nil, resp)
}
