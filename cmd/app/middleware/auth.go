package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/jwt"
	"github.com/HappyLadySauce/errors"
)

const (
	// UserIDKey is the key for user ID in context
	UserIDKey = "user_id"
	// UserRoleKey is the key for user role in context
	UserRoleKey = "user_role"
)

// JWTAuth creates a JWT authentication middleware.
func JWTAuth(s store.Factory) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()

		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			klog.V(1).Infof("missing authorization header: %s", authHeader)
			core.WriteResponse(c, errors.WithCode(code.ErrMissingHeader, "%s", code.Message(code.ErrMissingHeader)), nil)
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			klog.V(1).Infof("invalid authorization header format: %s", authHeader)
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "%s", code.Message(code.ErrInvalidAuthHeader)), nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析和验证 token
		claims, err := jwt.ParseToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			klog.V(1).Infof("invalid token: %v", err)
			// 检查是否是过期错误
			if strings.Contains(err.Error(), "expired") {
				core.WriteResponse(c, errors.WithCode(code.ErrExpired, "%s", code.Message(code.ErrExpired)), nil)
			} else {
				core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "%s", code.Message(code.ErrTokenInvalid)), nil)
			}
			c.Abort()
			return
		}

		if s == nil {
			klog.V(1).Infof("auth store is not initialized: %v", err)
			core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "%s", code.Message(code.ErrUnknown)), nil)
			c.Abort()
			return
		}

		// 查库校验用户状态/角色，确保注销/改角色立即生效
		user, err := s.Users().GetUser(context.Background(), claims.UserID)
		if err != nil {
			klog.V(1).Infof("failed to load user from store: %v", err)
			core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "%s", code.Message(code.ErrTokenInvalid)), nil)
			c.Abort()
			return
		}
		if user.Status != model.UserStatusActive {
			core.WriteResponse(c, errors.WithCode(code.ErrUserNotActive, "%s", code.Message(code.ErrUserNotActive)), nil)
			c.Abort()
			return
		}

		// 将用户信息存储到 context 中
		c.Set(UserIDKey, user.ID)
		c.Set(UserRoleKey, user.Role)

		c.Next()
	}
}
