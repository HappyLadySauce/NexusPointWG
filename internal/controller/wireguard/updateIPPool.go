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

// UpdateIPPool updates an IP pool by ID (admin only).
// @Summary Update IP pool
// @Description Update an IP pool by ID. Admin only.
// @Tags wireguard
// @Accept json
// @Produce json
// @Param id path string true "IP Pool ID"
// @Param pool body v1.UpdateIPPoolRequest true "IP pool update information"
// @Success 200 {object} v1.IPPoolResponse "IP pool updated successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - IP pool not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/ip-pools/{id} [put]
func (w *WGController) UpdateIPPool(c *gin.Context) {
	klog.V(1).Info("wireguard ip pool update function called.")

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
	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionIPPoolUpdate)
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
	var req v1.UpdateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Get existing IP pool
	existingPool, err := w.srv.IPPools().GetIPPool(context.Background(), poolID)
	if err != nil {
		klog.V(1).InfoS("failed to get IP pool", "poolID", poolID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// Check if CIDR is being modified
	if req.CIDR != nil && *req.CIDR != existingPool.CIDR {
		// Check if there are any allocated IPs from this pool
		hasAllocated, err := w.srv.IPPools().HasAllocatedIPs(context.Background(), poolID)
		if err != nil {
			klog.V(1).InfoS("failed to check allocated IPs", "poolID", poolID, "error", err)
			core.WriteResponse(c, err, nil)
			return
		}
		if hasAllocated {
			core.WriteResponse(c, errors.WithCode(code.ErrIPPoolInUse, "IP pool is in use and CIDR cannot be modified"), nil)
			return
		}
		// CIDR can be modified
		existingPool.CIDR = *req.CIDR
	}

	// Apply updates (only update provided fields)
	if req.Name != nil {
		existingPool.Name = *req.Name
	}
	if req.Routes != nil {
		existingPool.Routes = *req.Routes
	}
	if req.DNS != nil {
		existingPool.DNS = *req.DNS
	}
	if req.Endpoint != nil {
		existingPool.Endpoint = *req.Endpoint
	}
	if req.Description != nil {
		existingPool.Description = *req.Description
	}
	if req.Status != nil {
		existingPool.Status = *req.Status
	}

	// Update IP pool
	if err := w.srv.IPPools().UpdateIPPool(context.Background(), existingPool); err != nil {
		klog.V(1).InfoS("failed to update IP pool", "poolID", poolID, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	resp := v1.IPPoolResponse{
		ID:          existingPool.ID,
		Name:        existingPool.Name,
		CIDR:        existingPool.CIDR,
		Routes:      existingPool.Routes,
		DNS:         existingPool.DNS,
		Endpoint:    existingPool.Endpoint,
		Description: existingPool.Description,
		Status:      existingPool.Status,
		CreatedAt:   existingPool.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   existingPool.UpdatedAt.Format(time.RFC3339),
	}

	klog.V(1).InfoS("wireguard IP pool updated successfully", "poolID", poolID)
	core.WriteResponse(c, nil, resp)
}
