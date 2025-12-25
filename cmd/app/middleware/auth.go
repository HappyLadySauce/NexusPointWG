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

		// 规范化 Authorization header：去除多余空格，处理重复的 Bearer 前缀
		authHeader = strings.TrimSpace(authHeader)
		klog.V(2).Infof("received authorization header: [%s]", authHeader)

		// 移除所有可能的 Bearer 前缀（大小写不敏感），直到只剩下 token
		authHeaderLower := strings.ToLower(authHeader)
		for strings.HasPrefix(authHeaderLower, "bearer ") {
			// 找到第一个 "bearer " 的位置（不区分大小写）
			idx := strings.Index(authHeaderLower, "bearer ")
			if idx == 0 {
				// 从开头移除 "bearer "（保持原大小写）
				authHeader = strings.TrimSpace(authHeader[7:])
				authHeaderLower = strings.ToLower(authHeader)
			} else {
				break
			}
		}

		// 如果去除所有 Bearer 前缀后为空，说明格式错误
		if authHeader == "" {
			klog.Errorf("authorization header is empty after removing Bearer prefix")
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "%s", code.Message(code.ErrInvalidAuthHeader)), nil)
			c.Abort()
			return
		}

		// 现在 authHeader 应该只包含 token，重新添加标准的 Bearer 前缀
		authHeader = "Bearer " + strings.TrimSpace(authHeader)
		klog.V(2).Infof("normalized authorization header: [%s]", authHeader)

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			klog.Errorf("invalid authorization header format after normalization: [%s], length: %d, parts: %v", authHeader, len(authHeader), parts)
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
