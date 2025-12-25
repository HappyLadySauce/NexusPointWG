package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
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
)

// JWTAuth creates a JWT authentication middleware.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()

		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			klog.V(1).Info("missing authorization header")
			core.WriteResponse(c, errors.WithCode(code.ErrMissingHeader, "missing authorization header"), nil)
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			klog.V(1).Info("invalid authorization header format")
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "invalid authorization header format"), nil)
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
				core.WriteResponse(c, errors.WithCode(code.ErrExpired, "token expired"), nil)
			} else {
				core.WriteResponse(c, errors.WithCode(code.ErrTokenInvalid, "invalid token"), nil)
			}
			c.Abort()
			return
		}

		// 将用户信息存储到 context 中
		c.Set(UserIDKey, claims.UserID)

		c.Next()
	}
}
