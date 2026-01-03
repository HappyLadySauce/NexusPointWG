package wireguard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/network"
	"github.com/HappyLadySauce/errors"
)

// DownloadPeerConfig downloads the WireGuard client configuration file for a peer.
// @Summary Download WireGuard peer configuration
// @Description Download the WireGuard client configuration file for a peer. Admin can download any peer's config, regular users can only download their own peer configs.
// @Tags wireguard
// @Produce text/plain
// @Param id path string true "Peer ID"
// @Success 200 {string} string "Configuration file content"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid peer ID"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid or expired token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied"
// @Failure 404 {object} core.ErrResponse "Not found - peer not found or config file not found"
// @Failure 500 {object} core.ErrResponse "Internal server error"
// @Router /api/v1/wg/peers/{id}/config [get]
func (w *WGController) DownloadPeerConfig(c *gin.Context) {
	klog.V(1).Info("wireguard peer config download function called.")

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
	obj := spec.Obj(spec.ResourceWGConfig, scope)

	allowed, err := spec.Enforce(requesterRole, obj, spec.ActionWGConfigDownload)
	if err != nil {
		klog.V(1).InfoS("authz enforce failed", "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
		return
	}
	if !allowed {
		core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
		return
	}

	// Get WireGuard config
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		core.WriteResponse(c, errors.WithCode(code.ErrWGConfigNotInitialized, "WireGuard config not initialized"), nil)
		return
	}

	wgOpts := cfg.WireGuard

	// Try to read existing config file first
	userDir := wgOpts.ResolvedUserDir()
	configPath := filepath.Join(userDir, peerID+".conf")

	configContent, err := os.ReadFile(configPath)
	if err == nil {
		// File exists, return it
		c.Data(200, "text/plain; charset=utf-8", configContent)
		return
	}

	// File doesn't exist, generate it on the fly
	// Get server public key
	var serverPublicKey string
	if wgOpts.ServerConfigPath() != "" {
		configManager := wireguard.NewServerConfigManager(wgOpts.ServerConfigPath(), wgOpts.ApplyMethod)
		serverPublicKey, err = configManager.GetServerPublicKey()
		if err != nil {
			klog.V(1).InfoS("failed to get server public key", "error", err)
			// Continue with empty server public key
		}
	}

	// Get IP pool configuration if peer has IPPoolID
	var pool *model.IPPool
	if peer.IPPoolID != "" {
		var err error
		pool, err = w.srv.IPPools().GetIPPool(context.Background(), peer.IPPoolID)
		if err != nil {
			klog.V(1).InfoS("failed to get IP pool", "poolID", peer.IPPoolID, "error", err)
			// Continue without pool config
		}
	}

	// Use defaults if peer fields are empty
	// Priority: Peer specified > IP Pool config > Global config
	// If all are empty, DNS will be empty string and won't be written to config
	dns := peer.DNS
	if dns == "" && pool != nil && pool.DNS != "" {
		dns = pool.DNS
	}
	if dns == "" && wgOpts.DNS != "" {
		dns = wgOpts.DNS
	}
	// If dns is still empty, it will be omitted in GenerateClientConfig

	// Build endpoint with priority: Peer.Endpoint > Pool.Endpoint > Settings.ServerIP:ListenPort
	endpoint := peer.Endpoint
	if endpoint == "" && pool != nil && pool.Endpoint != "" {
		endpoint = pool.Endpoint
	}
	if endpoint == "" {
		// Use Settings.ServerIP + ListenPort from server config
		if wgOpts.ServerConfigPath() != "" {
			configManager := wireguard.NewServerConfigManager(wgOpts.ServerConfigPath(), wgOpts.ApplyMethod)
			serverConfig, err := configManager.ReadServerConfig()
			if err == nil && serverConfig != nil && serverConfig.Interface != nil {
				serverIP := wgOpts.ServerIP
				if serverIP == "" {
					// Auto-detect if not set
					detectedIP, err := network.GetServerIP(context.Background(), "")
					if err == nil {
						serverIP = detectedIP
					}
				}
				if serverIP != "" && serverConfig.Interface.ListenPort > 0 {
					endpoint = fmt.Sprintf("%s:%d", serverIP, serverConfig.Interface.ListenPort)
				}
			}
		}
		// Fallback to global endpoint if still empty
		if endpoint == "" {
			endpoint = wgOpts.Endpoint
		}
	}

	allowedIPs := peer.AllowedIPs
	if allowedIPs == "" && pool != nil && pool.Routes != "" {
		allowedIPs = pool.Routes
	}
	if allowedIPs == "" {
		allowedIPs = wgOpts.DefaultAllowedIPs
	}

	// Generate client config
	clientConfig := &wireguard.ClientConfig{
		PrivateKey:          peer.ClientPrivateKey,
		Address:             peer.ClientIP,
		DNS:                 dns,
		Endpoint:            endpoint,
		PublicKey:           serverPublicKey,
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: peer.PersistentKeepalive,
	}

	configContentStr := wireguard.GenerateClientConfig(clientConfig)

	// Return config content
	c.Data(200, "text/plain; charset=utf-8", []byte(configContentStr))
}
