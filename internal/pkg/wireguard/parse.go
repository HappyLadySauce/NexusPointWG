package wireguard

import (
	"encoding/json"
	"strconv"
	"strings"
)

// ParseServerConfig parses a server configuration file from []byte.
func ParseServerConfig(data []byte) (*ServerConfig, error) {
	config := &ServerConfig{
		Peers:    make([]*PeerBlock, 0),
		RawLines: make([]string, 0),
	}

	lines := strings.Split(string(data), "\n")
	var currentInterface *InterfaceBlock
	var currentPeer *PeerBlock
	var inInterface bool
	var inPeer bool
	var peerComment string
	var inManagedBlock bool // Track if we're inside the managed block

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for managed block markers
		if trimmed == ManagedBlockBegin {
			inManagedBlock = true
			// Keep the marker in RawLines for rendering
			if !inInterface && !inPeer {
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}
		if trimmed == ManagedBlockEnd {
			inManagedBlock = false
			// Keep the marker in RawLines for rendering
			if !inInterface && !inPeer {
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}

		// Skip empty lines (but preserve them in RawLines if not in a block)
		if trimmed == "" {
			if !inInterface && !inPeer {
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}

		// Check for section headers
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.Trim(trimmed, "[]")
			section = strings.TrimSpace(section)

			// Save previous block
			if inInterface && currentInterface != nil {
				config.Interface = currentInterface
				currentInterface = nil
			}
			if inPeer && currentPeer != nil {
				// Mark peer as managed if we're in managed block or comment indicates it
				if inManagedBlock || (peerComment != "" && strings.Contains(peerComment, "# NPWG peer_id=")) {
					currentPeer.IsManaged = true
				}
				config.Peers = append(config.Peers, currentPeer)
				currentPeer = nil
			}

			// Start new block
			inInterface = false
			inPeer = false

			if strings.EqualFold(section, "Interface") {
				inInterface = true
				currentInterface = &InterfaceBlock{
					Extra: make(map[string]string),
				}
				// Check if previous line was a comment
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if strings.HasPrefix(prevLine, "#") {
						// Comment belongs to interface, but we don't store it separately
					}
				}
			} else if strings.EqualFold(section, "Peer") {
				inPeer = true
				currentPeer = &PeerBlock{
					Extra:     make(map[string]string),
					IsManaged: inManagedBlock, // Mark as managed if we're in managed block
				}
				// Check if previous line was a comment
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if strings.HasPrefix(prevLine, "#") {
						peerComment = prevLine
						// Check if comment indicates managed peer
						if strings.Contains(peerComment, "# NPWG peer_id=") {
							currentPeer.IsManaged = true
						}
					}
				}
				if peerComment != "" {
					currentPeer.Comment = peerComment
					peerComment = ""
				}
			} else {
				// Unknown section, preserve as raw line
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}

		// Check for comments
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			if inPeer {
				// Comment before peer block
				peerComment = trimmed
				// Check if this is a managed peer comment
				if strings.Contains(trimmed, "# NPWG peer_id=") && currentPeer != nil {
					currentPeer.IsManaged = true
				}
			} else if !inInterface && !inPeer {
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}

		// Parse key-value pairs
		key, val, ok := splitKV(trimmed)
		if !ok {
			if !inInterface && !inPeer {
				config.RawLines = append(config.RawLines, line)
			}
			continue
		}

		keyLower := strings.ToLower(key)

		if inInterface && currentInterface != nil {
			switch keyLower {
			case "privatekey":
				currentInterface.PrivateKey = val
			case "address":
				currentInterface.Address = val
			case "mtu":
				currentInterface.MTU = val
			case "dns":
				currentInterface.DNS = val
			case "listenport":
				currentInterface.ListenPort = val
			case "postup":
				currentInterface.PostUp = val
			case "postdown":
				currentInterface.PostDown = val
			default:
				currentInterface.Extra[key] = val
			}
		} else if inPeer && currentPeer != nil {
			switch keyLower {
			case "publickey":
				currentPeer.PublicKey = val
			case "allowedips":
				currentPeer.AllowedIPs = val
			case "endpoint":
				currentPeer.Endpoint = val
			case "persistentkeepalive":
				// Try to parse as int, but store as string in Extra if needed
				currentPeer.PersistentKeepalive = parseIntOrDefault(val, 0)
			default:
				currentPeer.Extra[key] = val
			}
		} else {
			// Not in any block, preserve as raw line
			config.RawLines = append(config.RawLines, line)
		}
	}

	// Save last block if exists
	if inInterface && currentInterface != nil {
		config.Interface = currentInterface
	}
	if inPeer && currentPeer != nil {
		// Mark peer as managed if we're in managed block or comment indicates it
		if inManagedBlock || (currentPeer.Comment != "" && strings.Contains(currentPeer.Comment, "# NPWG peer_id=")) {
			currentPeer.IsManaged = true
		}
		config.Peers = append(config.Peers, currentPeer)
	}

	return config, nil
}

// ParseClientConfig parses a client configuration file (peer.conf) from []byte.
func ParseClientConfig(data []byte) (*ClientConfig, error) {
	config := &ClientConfig{}

	lines := strings.Split(string(data), "\n")
	var currentInterface *InterfaceBlock
	var currentPeer *PeerBlock
	var inInterface bool
	var inPeer bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.Trim(trimmed, "[]")
			section = strings.TrimSpace(section)

			if strings.EqualFold(section, "Interface") {
				inInterface = true
				inPeer = false
				currentInterface = &InterfaceBlock{
					Extra: make(map[string]string),
				}
			} else if strings.EqualFold(section, "Peer") {
				inPeer = true
				inInterface = false
				currentPeer = &PeerBlock{
					Extra: make(map[string]string),
				}
			}
			continue
		}

		// Skip comments
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		// Parse key-value pairs
		key, val, ok := splitKV(trimmed)
		if !ok {
			continue
		}

		keyLower := strings.ToLower(key)

		if inInterface && currentInterface != nil {
			switch keyLower {
			case "privatekey":
				currentInterface.PrivateKey = val
			case "address":
				currentInterface.Address = val
			case "mtu":
				currentInterface.MTU = val
			case "dns":
				currentInterface.DNS = val
			default:
				currentInterface.Extra[key] = val
			}
		} else if inPeer && currentPeer != nil {
			switch keyLower {
			case "publickey":
				currentPeer.PublicKey = val
			case "allowedips":
				currentPeer.AllowedIPs = val
			case "endpoint":
				currentPeer.Endpoint = val
			case "persistentkeepalive":
				currentPeer.PersistentKeepalive = parseIntOrDefault(val, 0)
			default:
				currentPeer.Extra[key] = val
			}
		}
	}

	config.Interface = currentInterface
	config.Peer = currentPeer

	return config, nil
}

// RenderServerConfig renders a ServerConfig to []byte.
func RenderServerConfig(config *ServerConfig) []byte {
	if config == nil {
		return nil
	}

	var b strings.Builder

	// Write raw lines before Interface block (excluding managed block markers)
	for _, line := range config.RawLines {
		trimmed := strings.TrimSpace(line)
		// Skip managed block markers in RawLines, we'll insert them at the right place
		if trimmed == ManagedBlockBegin || trimmed == ManagedBlockEnd {
			continue
		}
		if !strings.Contains(line, "[Interface]") && !strings.Contains(line, "[Peer]") {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Write Interface block
	if config.Interface != nil {
		b.WriteString("[Interface]\n")
		if config.Interface.PrivateKey != "" {
			b.WriteString("PrivateKey = ")
			b.WriteString(strings.TrimSpace(config.Interface.PrivateKey))
			b.WriteString("\n")
		}
		if config.Interface.Address != "" {
			b.WriteString("Address = ")
			b.WriteString(strings.TrimSpace(config.Interface.Address))
			b.WriteString("\n")
		}
		if config.Interface.ListenPort != "" {
			b.WriteString("ListenPort = ")
			b.WriteString(strings.TrimSpace(config.Interface.ListenPort))
			b.WriteString("\n")
		}
		if config.Interface.MTU != "" {
			b.WriteString("MTU = ")
			b.WriteString(strings.TrimSpace(config.Interface.MTU))
			b.WriteString("\n")
		}
		if config.Interface.DNS != "" {
			b.WriteString("DNS = ")
			b.WriteString(strings.TrimSpace(config.Interface.DNS))
			b.WriteString("\n")
		}
		if config.Interface.PostUp != "" {
			b.WriteString("PostUp   = ")
			b.WriteString(strings.TrimSpace(config.Interface.PostUp))
			b.WriteString("\n")
		}
		if config.Interface.PostDown != "" {
			b.WriteString("PostDown = ")
			b.WriteString(strings.TrimSpace(config.Interface.PostDown))
			b.WriteString("\n")
		}
		// Write extra fields
		for k, v := range config.Interface.Extra {
			b.WriteString(k)
			b.WriteString(" = ")
			b.WriteString(strings.TrimSpace(v))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Write Peer blocks with managed block markers
	hasManagedPeers := false
	for i, peer := range config.Peers {
		if peer == nil {
			continue
		}

		// Check if this is the first managed peer
		if peer.IsManaged && !hasManagedPeers {
			// Insert managed block begin marker before first managed peer
			b.WriteString(ManagedBlockBegin)
			b.WriteString("\n")
			hasManagedPeers = true
		}

		// Check if this is the last managed peer
		isLastManagedPeer := false
		if peer.IsManaged {
			// Check if there are more managed peers after this one
			hasMoreManaged := false
			for j := i + 1; j < len(config.Peers); j++ {
				if config.Peers[j] != nil && config.Peers[j].IsManaged {
					hasMoreManaged = true
					break
				}
			}
			if !hasMoreManaged {
				isLastManagedPeer = true
			}
		}

		// Write peer block
		if peer.Comment != "" {
			b.WriteString(peer.Comment)
			b.WriteString("\n")
		}
		b.WriteString("[Peer]\n")
		if peer.PublicKey != "" {
			b.WriteString("PublicKey = ")
			b.WriteString(strings.TrimSpace(peer.PublicKey))
			b.WriteString("\n")
		}
		if peer.AllowedIPs != "" {
			b.WriteString("AllowedIPs = ")
			b.WriteString(strings.TrimSpace(peer.AllowedIPs))
			b.WriteString("\n")
		}
		if peer.Endpoint != "" {
			b.WriteString("Endpoint = ")
			b.WriteString(strings.TrimSpace(peer.Endpoint))
			b.WriteString("\n")
		}
		if peer.PersistentKeepalive > 0 {
			b.WriteString("PersistentKeepalive = ")
			b.WriteString(strconv.Itoa(peer.PersistentKeepalive))
			b.WriteString("\n")
		}
		// Write extra fields
		for k, v := range peer.Extra {
			b.WriteString(k)
			b.WriteString(" = ")
			b.WriteString(strings.TrimSpace(v))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// Insert managed block end marker after last managed peer
		if isLastManagedPeer {
			b.WriteString(ManagedBlockEnd)
			b.WriteString("\n")
		}
	}

	// If we had managed peers but didn't close the block (shouldn't happen, but safety check)
	if hasManagedPeers {
		// Check if we already wrote the end marker
		output := b.String()
		if !strings.Contains(output, ManagedBlockEnd) {
			b.WriteString(ManagedBlockEnd)
			b.WriteString("\n")
		}
	}

	return []byte(b.String())
}

// RenderClientConfig renders a ClientConfig to []byte.
func RenderClientConfig(config *ClientConfig) []byte {
	if config == nil {
		return nil
	}

	var b strings.Builder

	// Write Interface block
	if config.Interface != nil {
		b.WriteString("[Interface]\n")
		if config.Interface.PrivateKey != "" {
			b.WriteString("PrivateKey = ")
			b.WriteString(strings.TrimSpace(config.Interface.PrivateKey))
			b.WriteString("\n")
		}
		if config.Interface.Address != "" {
			b.WriteString("Address = ")
			b.WriteString(strings.TrimSpace(config.Interface.Address))
			b.WriteString("\n")
		}
		if config.Interface.MTU != "" {
			b.WriteString("MTU = ")
			b.WriteString(strings.TrimSpace(config.Interface.MTU))
			b.WriteString("\n")
		}
		if config.Interface.DNS != "" {
			b.WriteString("DNS = ")
			b.WriteString(strings.TrimSpace(config.Interface.DNS))
			b.WriteString("\n")
		}
		// Write extra fields
		for k, v := range config.Interface.Extra {
			b.WriteString(k)
			b.WriteString(" = ")
			b.WriteString(strings.TrimSpace(v))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Write Peer block
	if config.Peer != nil {
		b.WriteString("[Peer]\n")
		if config.Peer.PublicKey != "" {
			b.WriteString("PublicKey = ")
			b.WriteString(strings.TrimSpace(config.Peer.PublicKey))
			b.WriteString("\n")
		}
		if config.Peer.AllowedIPs != "" {
			b.WriteString("AllowedIPs = ")
			b.WriteString(strings.TrimSpace(config.Peer.AllowedIPs))
			b.WriteString("\n")
		}
		if config.Peer.Endpoint != "" {
			b.WriteString("Endpoint = ")
			b.WriteString(strings.TrimSpace(config.Peer.Endpoint))
			b.WriteString("\n")
		}
		if config.Peer.PersistentKeepalive > 0 {
			b.WriteString("PersistentKeepalive = ")
			b.WriteString(strconv.Itoa(config.Peer.PersistentKeepalive))
			b.WriteString("\n")
		}
		// Write extra fields
		for k, v := range config.Peer.Extra {
			b.WriteString(k)
			b.WriteString(" = ")
			b.WriteString(strings.TrimSpace(v))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return []byte(b.String())
}

// RenderClientFiles renders all client files to a map of filename -> content.
func RenderClientFiles(files *ClientFiles) map[string][]byte {
	result := make(map[string][]byte)

	if files == nil {
		return result
	}

	// Render peer.conf
	if files.Config != nil {
		result["peer.conf"] = RenderClientConfig(files.Config)
	}

	// Render privatekey
	if files.PrivateKey != "" {
		result["privatekey"] = []byte(strings.TrimSpace(files.PrivateKey) + "\n")
	}

	// Render publickey
	if files.PublicKey != "" {
		result["publickey"] = []byte(strings.TrimSpace(files.PublicKey) + "\n")
	}

	// Render meta.json
	if files.Meta != nil {
		metaJSON, err := json.MarshalIndent(files.Meta, "", "  ")
		if err == nil {
			result["meta.json"] = append(metaJSON, '\n')
		}
	}

	return result
}

// ParseClientFiles parses all client configuration files from their raw content.
// It accepts a map of filename -> file content (as []byte).
func ParseClientFiles(files map[string][]byte) (*ClientFiles, error) {
	result := &ClientFiles{}

	// Parse peer.conf
	if peerConfData, ok := files["peer.conf"]; ok {
		config, err := ParseClientConfig(peerConfData)
		if err != nil {
			return nil, err
		}
		result.Config = config
	}

	// Parse privatekey
	if privKeyData, ok := files["privatekey"]; ok {
		result.PrivateKey = strings.TrimSpace(string(privKeyData))
	}

	// Parse publickey
	if pubKeyData, ok := files["publickey"]; ok {
		result.PublicKey = strings.TrimSpace(string(pubKeyData))
	}

	// Parse meta.json
	if metaData, ok := files["meta.json"]; ok {
		var meta ClientMeta
		if err := json.Unmarshal(metaData, &meta); err == nil {
			result.Meta = &meta
		}
		// If parsing fails, we just leave Meta as nil (optional field)
	}

	return result, nil
}

// Helper functions

func splitKV(line string) (k, v string, ok bool) {
	// WireGuard uses "Key = Value" but is tolerant on spaces.
	if !strings.Contains(line, "=") {
		return "", "", false
	}
	parts := strings.SplitN(line, "=", 2)
	k = strings.TrimSpace(parts[0])
	v = strings.TrimSpace(parts[1])
	if k == "" {
		return "", "", false
	}
	return k, v, true
}

func parseIntOrDefault(s string, defaultValue int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultValue
	}
	// Use strconv for proper integer parsing
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
