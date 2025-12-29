package v1

// CreateWGPeerRequest represents an admin peer creation request.
// swagger:model
type CreateWGPeerRequest struct {
	// Username is the owner username.
	Username string `json:"username" binding:"required,min=3,max=32,urlsafe,nochinese"`
	// DeviceName is a human friendly device name.
	DeviceName string `json:"deviceName" binding:"required,min=1,max=64"`
	// AllowedIPs optionally overrides server-side AllowedIPs for this peer.
	AllowedIPs string `json:"allowedIPs,omitempty" binding:"omitempty,max=512"`
	// PersistentKeepalive is optional keepalive in seconds (0 means unset).
	PersistentKeepalive *int `json:"persistentKeepalive,omitempty" binding:"omitempty,min=0,max=3600"`
}

// UpdateWGPeerRequest represents an admin peer update request.
// swagger:model
type UpdateWGPeerRequest struct {
	AllowedIPs          *string `json:"allowedIPs,omitempty" binding:"omitempty,max=512"`
	PersistentKeepalive *int    `json:"persistentKeepalive,omitempty" binding:"omitempty,min=0,max=3600"`
	Status              *string `json:"status,omitempty" binding:"omitempty,oneof=active revoked"`
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
	PersistentKeepalive int    `json:"persistent_keepalive"`
	Status              string `json:"status"`
}

// WGPeerListResponse represents a peer list response.
// swagger:model
type WGPeerListResponse struct {
	Total int64            `json:"total"`
	Items []WGPeerResponse `json:"items"`
}
