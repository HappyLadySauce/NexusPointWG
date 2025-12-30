package wireguard

// InterfaceBlock represents a [Interface] block in WireGuard configuration.
type InterfaceBlock struct {
	PrivateKey string
	Address    string
	MTU        string
	DNS        string
	ListenPort string
	PostUp     string
	PostDown   string
	// Additional fields can be stored in Extra map for unknown keys
	Extra map[string]string
}

// PeerBlock represents a [Peer] block in WireGuard configuration.
type PeerBlock struct {
	PublicKey           string
	AllowedIPs          string
	Endpoint            string
	PersistentKeepalive int
	Comment             string // Comment line before this peer block (e.g., "# home Ubuntu" or "# NPWG peer_id=...")
	IsManaged           bool   // Mark whether this peer belongs to the managed block (identified by comment "# NPWG peer_id=")
	// Additional fields can be stored in Extra map for unknown keys
	Extra map[string]string
}

// ServerConfig represents a complete server configuration file.
type ServerConfig struct {
	Interface *InterfaceBlock
	Peers     []*PeerBlock
	// RawLines stores lines that are not part of [Interface] or [Peer] blocks
	RawLines []string
}

// ClientConfig represents a complete client configuration file (peer.conf).
type ClientConfig struct {
	Interface *InterfaceBlock
	Peer      *PeerBlock
}

// ClientMeta represents the meta.json file structure.
type ClientMeta struct {
	PeerID      string `json:"peer_id"`
	User        string `json:"user"`
	DeviceName  string `json:"device_name"`
	ClientIP    string `json:"client_ip"`
	Endpoint    string `json:"endpoint"`
	GeneratedAt string `json:"generated_at"`
}

// ClientFiles represents all client configuration files.
type ClientFiles struct {
	Config     *ClientConfig
	PrivateKey string      // privatekey file content (trimmed)
	PublicKey  string      // publickey file content (trimmed)
	Meta       *ClientMeta // meta.json parsed content (optional)
}
