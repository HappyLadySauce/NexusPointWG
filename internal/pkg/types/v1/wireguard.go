package v1

// CreateWGPeerRequest represents an admin peer creation request.
// swagger:model
type CreateWGPeerRequest struct {
	// Username is the owner username.
	Username string `json:"username" binding:"required,min=3,max=32,urlsafe,nochinese"`
	// DeviceName is a human friendly device name.
	DeviceName string `json:"deviceName" binding:"required,min=1,max=64"`
	// AllowedIPs optionally overrides server-side AllowedIPs for this peer.
	AllowedIPs string `json:"allowedIPs,omitempty" binding:"omitempty,cidr,max=512"`
	// PersistentKeepalive is optional keepalive in seconds (0 means unset).
	PersistentKeepalive *int `json:"persistentKeepalive,omitempty" binding:"omitempty,min=0,max=3600"`
	// Endpoint optionally overrides the default endpoint from config.
	Endpoint *string `json:"endpoint,omitempty" binding:"omitempty,endpoint,max=255"`
	// DNS optionally overrides the default DNS from config.
	DNS *string `json:"dns,omitempty" binding:"omitempty,max=255"`
	// PrivateKey is an optional client private key. If not provided, one will be auto-generated.
	PrivateKey *string `json:"privateKey,omitempty" binding:"omitempty"`
}

// UpdateWGPeerRequest represents an admin peer update request.
// swagger:model
type UpdateWGPeerRequest struct {
	AllowedIPs          *string `json:"allowedIPs,omitempty" binding:"omitempty,cidr,max=512"`
	PersistentKeepalive *int    `json:"persistentKeepalive,omitempty" binding:"omitempty,min=0,max=3600"`
	DNS                 *string `json:"dns,omitempty" binding:"omitempty,max=255"`
	Status              *string `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
	// PrivateKey is an optional client private key. If provided, public key will be auto-generated.
	PrivateKey *string `json:"privateKey,omitempty" binding:"omitempty"`
	// Endpoint optionally overrides the default endpoint from config.
	Endpoint *string `json:"endpoint,omitempty" binding:"omitempty,endpoint,max=255"`
	// DeviceName is a human friendly device name.
	DeviceName *string `json:"deviceName,omitempty" binding:"omitempty,min=1,max=64"`
	// ClientIP is the client IP address in CIDR format (e.g., 100.100.100.5/32).
	ClientIP *string `json:"clientIP,omitempty" binding:"omitempty,cidr"`
}

// UserUpdateConfigRequest represents a user config update request.
// Note: User cannot update PrivateKey, DeviceName, Status, ClientIP.
// swagger:model
type UserUpdateConfigRequest struct {
	// AllowedIPs: IP ranges that the client is allowed to access.
	AllowedIPs *string `json:"allowedIPs,omitempty" binding:"omitempty,cidr,max=512"`
	// PersistentKeepalive: keepalive interval in seconds.
	PersistentKeepalive *int `json:"persistentKeepalive,omitempty" binding:"omitempty,min=0,max=3600"`
	// DNS: DNS server address.
	DNS *string `json:"dns,omitempty" binding:"omitempty,max=255"`
	// Endpoint: server endpoint address.
	Endpoint *string `json:"endpoint,omitempty" binding:"omitempty,endpoint,max=255"`
	// Note: PrivateKey, DeviceName, Status, ClientIP are not allowed for user updates.
}

// WGPeerResponse represents a WireGuard peer response.
// swagger:model
type WGPeerResponse struct {
	ID                  string `json:"id"`
	UserID              string `json:"user_id"`
	Username            string `json:"username"`
	DeviceName          string `json:"device_name"`
	ClientPublicKey     string `json:"client_public_key"`
	ClientIP            string `json:"client_ip"`
	AllowedIPs          string `json:"allowed_ips"`
	DNS                 string `json:"dns"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	Status              string `json:"status"`
}

// WGPeerListResponse represents a peer list response.
// swagger:model
type WGPeerListResponse struct {
	Total int64            `json:"total"`
	Items []WGPeerResponse `json:"items"`
}
