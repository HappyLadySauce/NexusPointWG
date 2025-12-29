package wireguard

import (
	"sort"
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
			b.WriteString(intToString(p.PersistentKeepalive))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(ManagedBlockEnd)
	b.WriteString("\n")
	return b.String()
}
