package model

import "time"

// WGPeer represents a WireGuard peer managed by NexusPointWG.
// It is the single source of truth for peer metadata; generated config files are derived artifacts.
type WGPeer struct {
	ID string `json:"id" gorm:"primaryKey"`

	// UserID is the owner user id (foreign key relation is managed at app level).
	UserID string `json:"user_id" gorm:"index;not null"`

	// DeviceName is a human friendly identifier, e.g. "laptop-01".
	DeviceName string `json:"device_name" gorm:"index;not null"`

	// ClientPublicKey is the peer public key (used in server config).
	ClientPublicKey string `json:"client_public_key" gorm:"uniqueIndex;not null"`

	// ClientIP is the allocated client address in CIDR form, e.g. "100.100.100.5/32".
	ClientIP string `json:"client_ip" gorm:"uniqueIndex;not null"`

	// AllowedIPs overrides client allowed ips. If empty, server/global defaults may be used.
	AllowedIPs string `json:"allowed_ips" gorm:"not null;default:''"`

	// PersistentKeepalive is optional keepalive value in seconds; 0 means unset.
	PersistentKeepalive int `json:"persistent_keepalive" gorm:"not null;default:0"`

	// Status is the peer lifecycle state.
	Status string `json:"status" gorm:"index;not null"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	WGPeerStatusActive  = "active"
	WGPeerStatusRevoked = "disabled"
)
