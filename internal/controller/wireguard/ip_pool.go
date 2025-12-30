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
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

// CreateIPPool creates a new IP pool (admin only).
// @Summary Create IP pool
// @Description Create a new IP pool for WireGuard peer IP allocation. Admin only.
// @Tags wireguard
// @Accept json
// @Produce json
// @Param pool body v1.CreateIPPoolRequest true "IP pool information"
// @Success 200 {object} v1.IPPoolResponse "IP pool created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/ip-pools [post]
func (w *WGController) CreateIPPool(c *gin.Context) {
	klog.V(1).Info("wireguard ip pool create function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceIPPool, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionIPPoolCreate)
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
	var req v1.CreateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Generate pool ID
	poolID, err := snowflake.GenerateID()
	if err != nil {
		klog.V(1).InfoS("failed to generate pool ID", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "failed to generate pool ID"), nil)
		return
	}

	// Create pool model
	pool := &model.IPPool{
		ID:          poolID,
		Name:        req.Name,
		CIDR:        req.CIDR,
		Routes:      req.Routes,
		DNS:         req.DNS,
		Endpoint:    req.Endpoint,
		Description: req.Description,
		Status:      model.IPPoolStatusActive,
	}

	// Create IP pool
	if err := w.srv.IPPools().CreateIPPool(context.Background(), pool); err != nil {
		klog.V(1).InfoS("failed to create IP pool", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp := v1.IPPoolResponse{
		ID:          pool.ID,
		Name:        pool.Name,
		CIDR:        pool.CIDR,
		Routes:      pool.Routes,
		DNS:         pool.DNS,
		Endpoint:    pool.Endpoint,
		Description: pool.Description,
		Status:      pool.Status,
		CreatedAt:   pool.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   pool.UpdatedAt.Format(time.RFC3339),
	}

	klog.V(1).InfoS("wireguard IP pool created successfully", "poolID", poolID)
	core.WriteResponse(c, nil, resp)
}

// ListIPPools lists IP pools with pagination (admin only).
// @Summary List IP pools
// @Description List IP pools with optional filters and pagination. Admin only.
// @Tags wireguard
// @Produce json
// @Param status query string false "Filter by status (active/disabled)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 20, max: 200)"
// @Success 200 {object} v1.IPPoolListResponse "IP pools listed successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid parameters"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/ip-pools [get]
func (w *WGController) ListIPPools(c *gin.Context) {
	klog.V(1).Info("wireguard ip pool list function called.")

	// Get requester info from JWTAuth middleware
	requesterRoleAny, _ := c.Get(middleware.UserRoleKey)
	requesterRole, _ := requesterRoleAny.(string)

	// --- Authorization (Casbin) - Admin only ---
	obj := spec.Obj(spec.ResourceIPPool, spec.ScopeAny)
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionIPPoolList)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Build list options
	opt := store.IPPoolListOptions{
		Status: c.Query("status"),
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

	// List IP pools
	pools, total, err := w.srv.IPPools().ListIPPools(context.Background(), opt)
	if err != nil {
		klog.V(1).InfoS("failed to list IP pools", "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Convert to response format
	items := make([]v1.IPPoolResponse, 0, len(pools))
	for _, pool := range pools {
		items = append(items, v1.IPPoolResponse{
			ID:          pool.ID,
			Name:        pool.Name,
			CIDR:        pool.CIDR,
			Routes:      pool.Routes,
			DNS:         pool.DNS,
			Endpoint:    pool.Endpoint,
			Description: pool.Description,
			Status:      pool.Status,
			CreatedAt:   pool.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   pool.UpdatedAt.Format(time.RFC3339),
		})
	}

	resp := v1.IPPoolListResponse{
		Total: total,
		Items: items,
	}

	core.WriteResponse(c, nil, resp)
}

// GetAvailableIPs gets available IP addresses from an IP pool (admin only).
// @Summary Get available IPs
// @Description Get a list of available IP addresses from an IP pool. Admin only.
// @Tags wireguard
// @Produce json
// @Param id path string true "IP Pool ID"
// @Param limit query int false "Limit the number of IPs to return (default: 50, max: 200)"
// @Success 200 {object} v1.AvailableIPsResponse "Available IPs retrieved successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid parameters"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - IP pool not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/ip-pools/{id}/available-ips [get]
func (w *WGController) GetAvailableIPs(c *gin.Context) {
	klog.V(1).Info("wireguard available IPs get function called.")

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
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionIPPoolList)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Get IP pool
	pool, err := w.srv.IPPools().GetIPPool(context.Background(), poolID)
	if err != nil {
		klog.V(1).InfoS("failed to get IP pool", "poolID", poolID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Parse limit
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	// Get available IPs from Service layer
	availableIPs, err := w.srv.IPPools().GetAvailableIPs(context.Background(), poolID, limit)
	if err != nil {
		klog.V(1).InfoS("failed to get available IPs", "poolID", poolID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp := v1.AvailableIPsResponse{
		IPPoolID: poolID,
		CIDR:     pool.CIDR,
		IPs:      availableIPs,
		Total:    len(availableIPs),
	}

	core.WriteResponse(c, nil, resp)
}
