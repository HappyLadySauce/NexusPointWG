package model

import "time"

// WGPeer represents a WireGuard peer/device owned by a user.
// This is a minimal model to support peer lifecycle + authz ownership checks.
type WGPeer struct {
	ID         string `json:"id" gorm:"primaryKey"`
	UserID     string `json:"user_id" gorm:"index;not null"` // owner
	Name       string `json:"name" gorm:"not null" validate:"required,min=1,max=64"`
	AllowedIPs string `json:"allowed_ips" gorm:"not null" validate:"required,max=255"`
	// ClientPublicKey is optional for now; real WG management may introduce full key lifecycle later.
	ClientPublicKey string    `json:"client_public_key" gorm:"not null;default:''" validate:"max=128"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
