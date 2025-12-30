package wireguard

import (
	"fmt"
	"strings"
)

// ClientConfig represents a WireGuard client configuration.
type ClientConfig struct {
	PrivateKey          string
	Address             string // Client IP in CIDR format, e.g. "100.100.100.2/32"
	DNS                 string // Optional, comma-separated
	MTU                 int    // Optional, default 1420
	Endpoint            string // Server endpoint, e.g. "118.24.41.142:51820"
	PublicKey           string // Server public key
	AllowedIPs          string // Comma-separated CIDRs
	PersistentKeepalive int    // Optional, default 25
}

// GenerateClientConfig generates a WireGuard client configuration file content.
func GenerateClientConfig(config *ClientConfig) string {
	var sb strings.Builder

	// Interface section
	sb.WriteString("[Interface]\n")
	if config.PrivateKey != "" {
		sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.PrivateKey))
	}
	if config.Address != "" {
		sb.WriteString(fmt.Sprintf("Address = %s\n", config.Address))
	}
	if config.DNS != "" {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", config.DNS))
	}
	if config.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", config.MTU))
	}
	sb.WriteString("\n")

	// Peer section
	sb.WriteString("[Peer]\n")
	if config.PublicKey != "" {
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", config.PublicKey))
	}
	if config.Endpoint != "" {
		sb.WriteString(fmt.Sprintf("Endpoint = %s\n", config.Endpoint))
	}
	if config.AllowedIPs != "" {
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", config.AllowedIPs))
	}
	if config.PersistentKeepalive > 0 {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", config.PersistentKeepalive))
	}

	return sb.String()
}

// ServerPeerConfig represents a peer configuration block in the server configuration.
type ServerPeerConfig struct {
	PublicKey           string
	AllowedIPs          string // Comma-separated CIDRs
	PersistentKeepalive int    // Optional
	Comment             string // Optional comment
}

// FormatServerPeerBlock formats a peer configuration block for the server config.
func FormatServerPeerBlock(peer *ServerPeerConfig) string {
	var sb strings.Builder

	if peer.Comment != "" {
		sb.WriteString(fmt.Sprintf("# %s\n", peer.Comment))
	}
	sb.WriteString("[Peer]\n")
	if peer.PublicKey != "" {
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
	}
	if peer.AllowedIPs != "" {
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs))
	}
	if peer.PersistentKeepalive > 0 {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
	}
	sb.WriteString("\n")

	return sb.String()
}
