package wireguard

import (
	"strings"
)

type InterfaceConfig struct {
	PrivateKey string
	Address    string
	MTU        string
}

// ParseInterfaceConfig parses the first [Interface] block and extracts a few fields used by NexusPointWG.
func ParseInterfaceConfig(conf string) InterfaceConfig {
	var cfg InterfaceConfig
	lines := strings.Split(conf, "\n")

	inIface := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sec := strings.Trim(line, "[]")
			inIface = (strings.EqualFold(sec, "Interface"))
			continue
		}
		if !inIface {
			continue
		}
		key, val, ok := splitKV(line)
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "privatekey":
			cfg.PrivateKey = strings.TrimSpace(val)
		case "address":
			// keep raw (could be comma-separated); allocator will pick the first v4 CIDR.
			cfg.Address = strings.TrimSpace(val)
		case "mtu":
			cfg.MTU = strings.TrimSpace(val)
		}
	}
	return cfg
}

// ExtractAllowedIPs returns all AllowedIPs raw values found in the config (across all [Peer] blocks).
func ExtractAllowedIPs(conf string) []string {
	var out []string
	lines := strings.Split(conf, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, val, ok := splitKV(line)
		if !ok {
			continue
		}
		if strings.EqualFold(key, "AllowedIPs") {
			out = append(out, strings.TrimSpace(val))
		}
	}
	return out
}

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
