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
	// UsernameKey is the key for username in context
	UsernameKey = "username"
	// UserRoleKey is the key for user role in context
	UserRoleKey = "user_role"
)

// JWTAuth creates a JWT authentication middleware.
func JWTAuth(s store.Factory) gin.HandlerFunc {
	return func(c *gin.Context) {
		if s == nil {
			klog.V(1).Info("authorization store is not initialized")
			core.WriteResponse(c, errors.WithCode(code.ErrStoreNotInitialized, "%s", code.Message(code.ErrStoreNotInitialized)), nil)
			c.Abort()
			return
		}

		cfg := config.Get()

		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			klog.V(1).Info("missing authorization header")
			core.WriteResponse(c, errors.WithCode(code.ErrMissingHeader, "%s", code.Message(code.ErrMissingHeader)), nil)
			c.Abort()
			return
		}

		// 去除前后空格
		authHeader = strings.TrimSpace(authHeader)

		// 去掉 Bearer 前缀（不区分大小写），只保留 token
		authHeaderLower := strings.ToLower(authHeader)
		if strings.HasPrefix(authHeaderLower, "bearer ") {
			// 去掉 "Bearer " 前缀（7 个字符），保留原大小写的 token
			authHeader = strings.TrimSpace(authHeader[7:])
		}

		// 如果 token 为空，说明格式错误
		if authHeader == "" {
			klog.V(1).Info("invalid authorization header format")
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "%s", code.Message(code.ErrInvalidAuthHeader)), nil)
			c.Abort()
			return
		}

		tokenString := authHeader

		// 解析和验证 token
		claims, err := jwt.ParseToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			klog.V(1).InfoS("failed to parse token", "error", err)
			// 检查是否是过期错误
			if strings.Contains(err.Error(), "expired") {
				core.WriteResponse(c, errors.WithCode(code.ErrExpired, "%s", code.Message(code.ErrExpired)), nil)
			} else {
				core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "%s", code.Message(code.ErrTokenInvalid)), nil)
			}
			c.Abort()
			return
		}

		// 查库校验用户状态/角色，确保注销/改角色立即生效
		user, err := s.Users().GetUser(context.Background(), claims.UserID)
		if err != nil {
			klog.V(1).InfoS("failed to get user from store:", "error", err)
			// 在认证过程中，如果用户不存在，应该返回认证失败（401）而不是资源不存在（404）
			// 这样可以确保前端能够正确识别认证失败并重定向到登录页
			if errors.ParseCoder(err).Code() == code.ErrUserNotFound {
				core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "%s", code.Message(code.ErrTokenInvalid)), nil)
			} else {
			core.WriteResponse(c, err, nil)
			}
			c.Abort()
			return
		}
		if user.Status != model.UserStatusActive {
			klog.V(1).InfoS("user is not active", "username", user.Username)
			core.WriteResponse(c, errors.WithCode(code.ErrUserNotActive, "%s", code.Message(code.ErrUserNotActive)), nil)
			c.Abort()
			return
		}

		// 将用户信息存储到 context 中
		c.Set(UserIDKey, user.ID)
		c.Set(UsernameKey, claims.Username)
		c.Set(UserRoleKey, user.Role)

		c.Next()
	}
}
