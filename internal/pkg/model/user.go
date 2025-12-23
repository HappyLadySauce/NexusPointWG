package model

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username" validate:"required,min=3,max=32"`
	Email        string    `json:"email" validate:"required,email,max=255"`
	Salt         string    `json:"salt,omitempty" gorm:"column:salt"`                   // 盐值，如果为空则自动生成
	PasswordHash string    `json:"password_hash,omitempty" gorm:"column:password_hash"` // 密码哈希，如果为空则自动生成
	Status       string    `json:"status" validate:"required,oneof=active inactive"`
	CreatedAt    time.Time `json:"created_at"` // 由 GORM 自动设置
	UpdatedAt    time.Time `json:"updated_at"` // 由 GORM 自动设置
}

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)
