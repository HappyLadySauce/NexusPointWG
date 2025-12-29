package wg

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
)

// Keep swagger type references resolvable by swag.
var _ = v1.WGPeerListResponse{}

// ListMyConfigs lists current user's peers/configs.
// @Summary List my WireGuard configs
// @Tags wg
// @Produce json
// @Success 200 {object} v1.WGPeerListResponse "Configs listed successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/configs [get]
func (w *WGController) ListMyConfigs(c *gin.Context) {
	userID, err := requesterUserID(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	if err := enforce(c, spec.Obj(spec.ResourceWGPeer, spec.ScopeSelf), spec.ActionWGPeerList); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	peers, err := w.srv.WG().UserListPeers(context.Background(), userID)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	resp, err := w.srv.WG().ToWGPeerListResponse(context.Background(), peers, int64(len(peers)))
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}

// DownloadConfig downloads a peer config (.conf).
// @Summary Download WireGuard config
// @Tags wg
// @Produce application/octet-stream
// @Param id path string true "Peer ID"
// @Success 200 {string} string "WireGuard config file"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 404 {object} core.ErrResponse "Not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/configs/{id}/download [get]
func (w *WGController) DownloadConfig(c *gin.Context) {
	userID, err := requesterUserID(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	if err := enforce(c, spec.Obj(spec.ResourceWGConfig, spec.ScopeSelf), spec.ActionWGConfigDownload); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	id := c.Param("id")
	filename, content, err := w.srv.WG().UserDownloadConfig(context.Background(), userID, id)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "application/octet-stream", content)
}

// RotateConfig rotates keys for a peer.
// @Summary Rotate WireGuard config
// @Tags wg
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} core.SuccessResponse "Rotated successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/configs/{id}/rotate [post]
func (w *WGController) RotateConfig(c *gin.Context) {
	userID, err := requesterUserID(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	if err := enforce(c, spec.Obj(spec.ResourceWGConfig, spec.ScopeSelf), spec.ActionWGConfigRotate); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	id := c.Param("id")
	if err := w.srv.WG().UserRotateConfig(context.Background(), userID, id); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}

// UpdateConfig updates a user's WireGuard config.
// @Summary Update WireGuard config
// @Description User can update AllowedIPs, DNS, PersistentKeepalive, Endpoint. Cannot update PrivateKey, DeviceName, Status, ClientIP.
// @Tags wg
// @Accept json
// @Produce json
// @Param id path string true "Peer ID"
// @Param config body v1.UserUpdateConfigRequest true "Update config payload"
// @Success 200 {object} core.SuccessResponse "Config updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or forbidden fields"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/configs/{id} [put]
func (w *WGController) UpdateConfig(c *gin.Context) {
	userID, err := requesterUserID(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	if err := enforce(c, spec.Obj(spec.ResourceWGConfig, spec.ScopeSelf), spec.ActionWGConfigUpdate); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	var req v1.UserUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	id := c.Param("id")
	if err := w.srv.WG().UserUpdateConfig(context.Background(), userID, id, req); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}

// RevokeConfig revokes a peer/config (self).
// @Summary Revoke WireGuard config
// @Tags wg
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} core.SuccessResponse "Deleted successfully"
// @Failure 401 {object} core.ErrResponse "Unauthorized"
// @Failure 403 {object} core.ErrResponse "Forbidden"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/configs/{id}/revoke [post]
func (w *WGController) RevokeConfig(c *gin.Context) {
	userID, err := requesterUserID(c)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	if err := enforce(c, spec.Obj(spec.ResourceWGConfig, spec.ScopeSelf), spec.ActionWGConfigRevoke); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	id := c.Param("id")
	if err := w.srv.WG().UserRevokeConfig(context.Background(), userID, id); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}
