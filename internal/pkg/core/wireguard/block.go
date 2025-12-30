package wireguard

import (
	"sort"
	"strconv"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

const (
	ManagedBlockBegin = "# NexusPointWG BEGIN"
	ManagedBlockEnd   = "# NexusPointWG END"
)

// RenderManagedBlock renders the managed block section for wg0.conf.
// Only peers in this list will be emitted inside the managed block.
func RenderManagedBlock(peers []*model.WGPeer) string {
	var b strings.Builder
	b.WriteString(ManagedBlockBegin)
	b.WriteString("\n")

	// Stable order: by ClientIP then ID
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i] == nil && peers[j] == nil {
			return false
		}
		if peers[i] == nil {
			return false
		}
		if peers[j] == nil {
			return true
		}
		if peers[i].ClientIP == peers[j].ClientIP {
			return peers[i].ID < peers[j].ID
		}
		return peers[i].ClientIP < peers[j].ClientIP
	})

	for _, p := range peers {
		if p == nil {
			continue
		}
		// Skip disabled peers
		if strings.TrimSpace(p.Status) == model.WGPeerStatusDisabled {
			continue
		}
		// 服务器配置中的 AllowedIPs 始终使用客户端 IP
		// 这表示服务器允许该客户端访问的网络（即客户端的 IP 地址）
		allowed := strings.TrimSpace(p.ClientIP)
		if allowed == "" || strings.TrimSpace(p.ClientPublicKey) == "" {
			// Incomplete record; skip for safety.
			continue
		}

		b.WriteString("# NPWG peer_id=")
		b.WriteString(p.ID)
		b.WriteString(" user_id=")
		b.WriteString(p.UserID)
		b.WriteString(" device=")
		b.WriteString(p.DeviceName)
		b.WriteString("\n")
		b.WriteString("[Peer]\n")
		b.WriteString("PublicKey = ")
		b.WriteString(strings.TrimSpace(p.ClientPublicKey))
		b.WriteString("\n")
		b.WriteString("AllowedIPs = ")
		b.WriteString(allowed)
		b.WriteString("\n")
		if p.PersistentKeepalive > 0 {
			b.WriteString("PersistentKeepalive = ")
			b.WriteString(strconv.Itoa(p.PersistentKeepalive))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(ManagedBlockEnd)
	b.WriteString("\n")
	return b.String()
}

// PeerBlockFromModel converts a model.WGPeer to a PeerBlock for the managed block.
// The returned PeerBlock will have IsManaged = true and the appropriate comment.
func PeerBlockFromModel(peer *model.WGPeer) *PeerBlock {
	if peer == nil {
		return nil
	}

	// Generate comment: "# NPWG peer_id=... user_id=... device=..."
	var commentBuilder strings.Builder
	commentBuilder.WriteString("# NPWG peer_id=")
	commentBuilder.WriteString(peer.ID)
	commentBuilder.WriteString(" user_id=")
	commentBuilder.WriteString(peer.UserID)
	commentBuilder.WriteString(" device=")
	commentBuilder.WriteString(peer.DeviceName)

	// Server config AllowedIPs always uses client IP
	allowed := strings.TrimSpace(peer.ClientIP)

	return &PeerBlock{
		PublicKey:           strings.TrimSpace(peer.ClientPublicKey),
		AllowedIPs:          allowed,
		PersistentKeepalive: peer.PersistentKeepalive,
		Comment:             commentBuilder.String(),
		IsManaged:           true,
		Extra:               make(map[string]string),
	}
}

// ReplaceManagedBlockPeers replaces the managed block peers in ServerConfig.
// It preserves peers outside the managed block and replaces managed peers with new ones from DB.
func ReplaceManagedBlockPeers(config *ServerConfig, peers []*model.WGPeer) {
	if config == nil {
		return
	}

	// Separate managed and unmanaged peers
	unmanagedPeers := make([]*PeerBlock, 0)
	for _, peer := range config.Peers {
		if peer != nil && !peer.IsManaged {
			unmanagedPeers = append(unmanagedPeers, peer)
		}
	}

	// Convert model.WGPeer to PeerBlock for managed peers
	// Filter disabled peers and incomplete records
	managedPeers := make([]*PeerBlock, 0)
	for _, p := range peers {
		if p == nil {
			continue
		}
		// Skip disabled peers
		if strings.TrimSpace(p.Status) == model.WGPeerStatusDisabled {
			continue
		}
		// Skip incomplete records
		allowed := strings.TrimSpace(p.ClientIP)
		if allowed == "" || strings.TrimSpace(p.ClientPublicKey) == "" {
			continue
		}
		managedPeers = append(managedPeers, PeerBlockFromModel(p))
	}

	// Sort managed peers: by ClientIP then ID (same as RenderManagedBlock)
	sort.SliceStable(managedPeers, func(i, j int) bool {
		if managedPeers[i] == nil && managedPeers[j] == nil {
			return false
		}
		if managedPeers[i] == nil {
			return false
		}
		if managedPeers[j] == nil {
			return true
		}
		if managedPeers[i].AllowedIPs == managedPeers[j].AllowedIPs {
			// Extract peer_id from comment for comparison
			iID := extractPeerIDFromComment(managedPeers[i].Comment)
			jID := extractPeerIDFromComment(managedPeers[j].Comment)
			return iID < jID
		}
		return managedPeers[i].AllowedIPs < managedPeers[j].AllowedIPs
	})

	// Recombine: unmanaged peers + managed peers
	config.Peers = append(unmanagedPeers, managedPeers...)
}

// extractPeerIDFromComment extracts peer_id from comment like "# NPWG peer_id=xxx ..."
func extractPeerIDFromComment(comment string) string {
	if !strings.Contains(comment, "peer_id=") {
		return ""
	}
	// Find "peer_id=" and extract until next space
	start := strings.Index(comment, "peer_id=")
	if start == -1 {
		return ""
	}
	start += len("peer_id=")
	end := strings.Index(comment[start:], " ")
	if end == -1 {
		return comment[start:]
	}
	return comment[start : start+end]
}
