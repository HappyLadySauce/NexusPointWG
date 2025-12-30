package model

import (
	"time"
)

// WGPeer represents a WireGuard peer configuration.
type WGPeer struct {
	ID                  string    `json:"id" gorm:"primaryKey"`
	UserID              string    `json:"user_id" gorm:"index;not null"`
	DeviceName          string    `json:"device_name" gorm:"not null"`
	ClientPrivateKey    string    `json:"client_private_key,omitempty" gorm:"column:client_private_key;not null"`
	ClientPublicKey     string    `json:"client_public_key" gorm:"uniqueIndex;not null"`
	ClientIP            string    `json:"client_ip" gorm:"index;not null"` // IPv4 CIDR, e.g. "100.100.100.2/32"
	AllowedIPs          string    `json:"allowed_ips" gorm:"not null"`     // Comma-separated CIDRs
	DNS                 string    `json:"dns" gorm:""`                      // Optional, comma-separated
	Endpoint            string    `json:"endpoint" gorm:""`                // Optional, overrides server default
	PersistentKeepalive int       `json:"persistent_keepalive" gorm:"default:25"`
	Status              string    `json:"status" gorm:"not null;default:active"` // active, disabled
	IPPoolID            string    `json:"ip_pool_id" gorm:"index"`                // 关联的IP池
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

const (
	// WGPeerStatusActive indicates the peer is active and can be used.
	WGPeerStatusActive = "active"
	// WGPeerStatusDisabled indicates the peer is disabled and cannot be used.
	WGPeerStatusDisabled = "disabled"
)

