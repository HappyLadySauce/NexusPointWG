package model

import (
	"time"
)

type User struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=32,urlsafe,nochinese"`
	Nickname     string    `json:"nickname" gorm:"not null" validate:"required,min=3,max=32"`
	Avatar       string    `json:"avatar" gorm:"not null" validate:"required,url,max=255"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email,max=255,emaildomain"`
	Salt         string    `json:"salt,omitempty" gorm:"column:salt;not null"`                   // 盐值，如果为空则自动生成
	PasswordHash string    `json:"password_hash,omitempty" gorm:"column:password_hash;not null"` // 密码哈希，如果为空则自动生成
	Status       string    `json:"status" gorm:"not null" validate:"required,oneof=active inactive deleted"`
	Role         string    `json:"role" gorm:"not null" validate:"required,oneof=user admin"`
	CreatedAt    time.Time `json:"created_at"` // 由 GORM 自动设置
	UpdatedAt    time.Time `json:"updated_at"` // 由 GORM 自动设置
}

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusDeleted  = "deleted"
	UserRoleUser       = "user"
	UserRoleAdmin      = "admin"
	// DefaultAvatarURL is the default avatar URL for users who don't provide one
	DefaultAvatarURL = "https://image.happyladysauce.cn/img/2025/10/28/1/auther.ico"
)
