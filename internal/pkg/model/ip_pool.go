package model

import (
	"time"
)

// IPPool represents an IP address pool for WireGuard peer allocation.
type IPPool struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`             // 地址池名称
	CIDR        string    `json:"cidr" gorm:"column:cidr;uniqueIndex;not null"` // e.g. "100.100.100.0/24"
	Routes      string    `json:"routes" gorm:""`                               // 路由（逗号分隔的CIDR），用于客户端的AllowedIPs
	DNS         string    `json:"dns" gorm:""`                                  // DNS服务器（逗号分隔），用于客户端配置
	Endpoint    string    `json:"endpoint" gorm:""`                             // 服务器端点，格式如 "10.10.10.10:51820"
	Description string    `json:"description" gorm:""`                          // 描述
	Status      string    `json:"status" gorm:"not null;default:active"`        // active, disabled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	// IPPoolStatusActive indicates the IP pool is active and can be used for allocation.
	IPPoolStatusActive = "active"
	// IPPoolStatusDisabled indicates the IP pool is disabled and cannot be used for allocation.
	IPPoolStatusDisabled = "disabled"
)
