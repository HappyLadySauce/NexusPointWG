package wg

import (
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
)

// toWGPeerResponse converts a model.WGPeer and model.User to a v1.WGPeerResponse.
func toWGPeerResponse(peer *model.WGPeer, username string) *v1.WGPeerResponse {
	if peer == nil {
		return nil
	}

	// If AllowedIPs is empty, return the actual default value used
	allowedIPs := peer.AllowedIPs
	if strings.TrimSpace(allowedIPs) == "" {
		// Get default from config
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil {
			allowedIPs = strings.TrimSpace(cfg.WireGuard.DefaultAllowedIPs)
			if allowedIPs == "" {
				allowedIPs = "0.0.0.0/0,::/0" // Final default
			}
		} else {
			allowedIPs = "0.0.0.0/0,::/0" // Fallback default
		}
	}

	// If DNS is empty, return the actual default value used
	dns := peer.DNS
	if strings.TrimSpace(dns) == "" {
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil {
			dns = strings.TrimSpace(cfg.WireGuard.DNS)
		}
	}

	return &v1.WGPeerResponse{
		ID:                  peer.ID,
		UserID:              peer.UserID,
		Username:            username,
		DeviceName:          peer.DeviceName,
		ClientPublicKey:     peer.ClientPublicKey,
		ClientIP:            peer.ClientIP,
		AllowedIPs:          allowedIPs,
		DNS:                 dns,
		PersistentKeepalive: peer.PersistentKeepalive,
		Status:              peer.Status,
	}
}

// toWGPeerListResponse converts a list of model.WGPeer to a v1.WGPeerListResponse.
// It requires a map of userID -> username for username lookup.
func toWGPeerListResponse(peers []*model.WGPeer, total int64, userMap map[string]string) *v1.WGPeerListResponse {
	items := make([]v1.WGPeerResponse, 0, len(peers))
	for _, p := range peers {
		if p == nil {
			continue
		}
		username := userMap[p.UserID]
		if username == "" {
			username = "" // Fallback to empty if user not found
		}
		items = append(items, *toWGPeerResponse(p, username))
	}
	return &v1.WGPeerListResponse{
		Total: total,
		Items: items,
	}
}
