package v1

// CreateWGPeerRequest represents a peer creation request.
// swagger:model
type CreateWGPeerRequest struct {
	// Name is a human-readable device name.
	Name string `json:"name" binding:"required,min=1,max=64"`
	// AllowedIPs is the client AllowedIPs string (e.g. "10.0.0.2/32").
	AllowedIPs string `json:"allowed_ips" binding:"required,max=255"`
	// ClientPublicKey is optional in this minimal scaffold.
	ClientPublicKey string `json:"client_public_key,omitempty" binding:"omitempty,max=128"`
	// UserID is optional; only admin may set it to create peers for other users.
	UserID string `json:"user_id,omitempty" binding:"omitempty,max=64"`
}

// UpdateWGPeerRequest represents a peer update request.
// swagger:model
type UpdateWGPeerRequest struct {
	Name            *string `json:"name,omitempty" binding:"omitempty,min=1,max=64"`
	AllowedIPs      *string `json:"allowed_ips,omitempty" binding:"omitempty,max=255"`
	ClientPublicKey *string `json:"client_public_key,omitempty" binding:"omitempty,max=128"`
}

// WGPeerResponse is a peer response payload.
// swagger:model
type WGPeerResponse struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Name            string `json:"name"`
	AllowedIPs      string `json:"allowed_ips"`
	ClientPublicKey string `json:"client_public_key,omitempty"`
}

// WGConfigResponse is a minimal config download payload.
// swagger:model
type WGConfigResponse struct {
	PeerID string `json:"peer_id"`
	Config string `json:"config"`
}
