package model

import (
	"time"
)

// IPPool represents an IP address pool for WireGuard peer allocation.
type IPPool struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"` // 地址池名称
	CIDR        string    `json:"cidr" gorm:"uniqueIndex;not null"`   // e.g. "100.100.100.0/24"
	ServerIP    string    `json:"server_ip" gorm:"not null"`         // 服务器IP, e.g. "100.100.100.1/32"
	Gateway     string    `json:"gateway" gorm:""`                   // 网关地址（可选）
	Description string    `json:"description" gorm:""`               // 描述
	Status      string    `json:"status" gorm:"not null;default:active"` // active, disabled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	// IPPoolStatusActive indicates the IP pool is active and can be used for allocation.
	IPPoolStatusActive = "active"
	// IPPoolStatusDisabled indicates the IP pool is disabled and cannot be used for allocation.
	IPPoolStatusDisabled = "disabled"
)

