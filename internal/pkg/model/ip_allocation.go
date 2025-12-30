package model

import (
	"time"
)

// IPAllocation represents an IP address allocation record for a WireGuard peer.
type IPAllocation struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	IPPoolID  string    `json:"ip_pool_id" gorm:"index;not null"`
	PeerID    string    `json:"peer_id" gorm:"uniqueIndex;not null"` // 关联的Peer
	IPAddress string    `json:"ip_address" gorm:"uniqueIndex;not null"` // e.g. "100.100.100.2"
	Status    string    `json:"status" gorm:"not null;default:allocated"` // allocated, released
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	// IPAllocationStatusAllocated indicates the IP address is allocated to a peer.
	IPAllocationStatusAllocated = "allocated"
	// IPAllocationStatusReleased indicates the IP address has been released and is available for reuse.
	IPAllocationStatusReleased = "released"
)

