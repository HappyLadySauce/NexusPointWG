package wg

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
)

// ListPeers lists WireGuard peers (admin can list any; user can list self via /wg/configs).
// @Summary List WireGuard peers
// @Description Admin: list all peers
// @Tags wg
// @Produce json
// @Param user_id query string false "Filter by user id"
// @Param device_name query string false "Filter by device name (contains)"
// @Param client_ip query string false "Filter by client ip (exact)"
// @Param status query string false "Filter by status (active/disabled)"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {object} v1.WGPeerListResponse "Peers listed successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers [get]
func (w *WGController) ListPeers(c *gin.Context) {
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeAny), spec.ActionWGPeerList); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	offset, _ := strconv.Atoi(c.Query("offset"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	peers, total, err := w.srv.WG().AdminListPeers(context.Background(), store.WGPeerListOptions{
		UserID:     c.Query("user_id"),
		DeviceName: c.Query("device_name"),
		ClientIP:   c.Query("client_ip"),
		Status:     c.Query("status"),
		Offset:     offset,
		Limit:      limit,
	})
	if err != nil {
		klog.V(1).InfoS("failed to list peers", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp, err := w.srv.WG().ToWGPeerListResponse(context.Background(), peers, total)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

// CreatePeer creates a WireGuard peer for a user (admin only).
// @Summary Create WireGuard peer
// @Description Admin: create peer (auto generate keys/ip, write files, apply)
// @Tags wg
// @Accept json
// @Produce json
// @Param peer body v1.CreateWGPeerRequest true "Create peer payload"
// @Success 200 {object} v1.WGPeerResponse "Peer created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers [post]
func (w *WGController) CreatePeer(c *gin.Context) {
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeAny), spec.ActionWGPeerCreate); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	var req v1.CreateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	peer, err := w.srv.WG().AdminCreatePeer(context.Background(), req)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	resp, err := w.srv.WG().ToWGPeerResponse(context.Background(), peer)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

// GetPeer returns a peer by id (admin only).
// @Summary Get WireGuard peer
// @Tags wg
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} v1.WGPeerResponse "Peer retrieved successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 404 {object} core.ErrResponse "Not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [get]
func (w *WGController) GetPeer(c *gin.Context) {
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeAny), spec.ActionWGPeerRead); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	id := c.Param("id")
	peer, err := w.srv.WG().AdminGetPeer(context.Background(), id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	resp, err := w.srv.WG().ToWGPeerResponse(context.Background(), peer)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

// UpdatePeer updates a peer (admin only).
// @Summary Update WireGuard peer
// @Tags wg
// @Accept json
// @Produce json
// @Param id path string true "Peer ID"
// @Param peer body v1.UpdateWGPeerRequest true "Update peer payload"
// @Success 200 {object} v1.WGPeerResponse "Peer updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 404 {object} core.ErrResponse "Not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [put]
func (w *WGController) UpdatePeer(c *gin.Context) {
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeAny), spec.ActionWGPeerUpdate); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	var req v1.UpdateWGPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponseBindErr(c, err, nil)
		return
	}
	id := c.Param("id")

	peer, err := w.srv.WG().AdminUpdatePeer(context.Background(), id, req)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	resp, err := w.srv.WG().ToWGPeerResponse(context.Background(), peer)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

// DeletePeer revokes a peer (admin only).
// @Summary Delete (revoke) WireGuard peer
// @Tags wg
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} core.SuccessResponse "Peer deleted successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 404 {object} core.ErrResponse "Not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id} [delete]
func (w *WGController) DeletePeer(c *gin.Context) {
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeAny), spec.ActionWGPeerDelete); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	id := c.Param("id")
	if err := w.srv.WG().AdminRevokePeer(context.Background(), id); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}
