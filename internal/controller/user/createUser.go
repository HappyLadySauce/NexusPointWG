package user

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/jwt"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

// CreateUser create a user.
// Supports two modes:
// 1. Public registration (unauthenticated): Creates a regular user with role=user, status=active
// 2. Admin creation (authenticated admin): Can create users with custom role and status
// @Summary Create user
// @Description Create a user with username, nickname, avatar, email and password. Supports public registration (unauthenticated) and admin creation (authenticated). Role and status fields are only available for authenticated admin users.
// @Tags users
// @Accept json
// @Produce json
// @Param user body v1.CreateUserRequest true "User information. Role and status fields are only available for authenticated admin users."
// @Success 200 {object} core.SuccessResponse "User created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - encryption error or invalid token"
// @Failure 403 {object} core.ErrResponse "Forbidden - permission denied (authenticated non-admin users cannot create users)"
// @Failure 500 {object} core.ErrResponse "Internal server error - database error"
// @Router /api/v1/users [post]
func (u *UserController) CreateUser(c *gin.Context) {
	klog.V(1).Info("user create function called.")

	var httpUser v1.CreateUserRequest
	var user model.User

	if err := c.ShouldBindJSON(&httpUser); err != nil {
		klog.V(1).InfoS("invalid request body", "error", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// Check if user is authenticated (optional - this endpoint supports both authenticated and unauthenticated access)
	// Try to get from middleware first (if route is in authed group)
	requesterIDAny, isAuthenticated := c.Get("user_id")
	requesterRoleAny, _ := c.Get("user_role")
	requesterRole, _ := requesterRoleAny.(string)

	// If not set by middleware, try to parse token manually (for optional auth support)
	if !isAuthenticated {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			authHeader = strings.TrimSpace(authHeader)
			authHeaderLower := strings.ToLower(authHeader)
			if strings.HasPrefix(authHeaderLower, "bearer ") {
				tokenString := strings.TrimSpace(authHeader[7:])
				if tokenString != "" {
					cfg := config.Get()
					claims, err := jwt.ParseToken(tokenString, cfg.JWT.Secret)
					if err == nil {
						// Token is valid, verify user exists and is active
						requesterUser, err := u.srv.Users().GetUser(context.Background(), claims.UserID)
						if err == nil && requesterUser.Status == model.UserStatusActive {
							requesterIDAny = claims.UserID
							requesterRole = requesterUser.Role // Use role from DB, not from token
							isAuthenticated = true
							klog.V(1).InfoS("optional auth successful", "userID", claims.UserID, "role", requesterUser.Role)
						} else {
							klog.V(1).InfoS("optional auth failed - user not found or not active", "userID", claims.UserID, "error", err)
						}
					}
				}
			}
		}
	}

	// If authenticated, check permissions
	if isAuthenticated {
		// Only admins can create users when authenticated
		obj := spec.Obj(spec.ResourceUser, spec.ScopeAny)
		allowed, err := spec.Enforce(requesterRole, obj, spec.ActionUserCreate)
		if err != nil {
			klog.V(1).InfoS("authz enforce failed", "requesterRole", requesterRole, "error", err)
			core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "authorization engine error"), nil)
			return
		}
		if !allowed {
			klog.V(1).InfoS("permission denied for user creation", "requesterRole", requesterRole, "requesterID", requesterIDAny)
			core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied)), nil)
			return
		}

		// Admin can specify role and status
		if httpUser.Role != nil {
			user.Role = *httpUser.Role
		} else {
			user.Role = model.UserRoleUser
		}

		if httpUser.Status != nil {
			user.Status = *httpUser.Status
		} else {
			user.Status = model.UserStatusActive
		}
	} else {
		// Public registration: only regular users, ignore role/status if provided
		user.Role = model.UserRoleUser
		user.Status = model.UserStatusActive
		if httpUser.Role != nil || httpUser.Status != nil {
			klog.V(1).InfoS("role/status fields ignored for public registration", "username", httpUser.Username)
		}
	}

	// 如果提供了原始密码，生成盐值和密码哈希
	if httpUser.Password != "" {
		// 生成盐值
		salt, err := passwd.GenerateSalt()
		if err != nil {
			klog.V(1).InfoS("failed to generate salt", "username", httpUser.Username, "error", err)
			core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
			return
		}

		// 生成密码哈希
		passwordHash, err := passwd.HashPassword(httpUser.Password, salt)
		if err != nil {
			klog.V(1).InfoS("failed to hash password", "username", httpUser.Username, "error", err)
			core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, "%s", err.Error()), nil)
			return
		}

		// 设置盐值和密码哈希
		user.Salt = salt
		user.PasswordHash = passwordHash
	}

	// 生成用户 ID (雪花算法)
	userID, err := snowflake.GenerateID()
	if err != nil {
		klog.V(1).InfoS("failed to generate user ID", "username", httpUser.Username, "error", err)
		core.WriteResponse(c, errors.WithCode(code.ErrUnknown, "failed to generate user ID"), nil)
		return
	}
	user.ID = userID

	// 设置用户名、昵称、头像和邮箱
	user.Username = httpUser.Username
	// 如果没有提供昵称，则使用用户名
	if httpUser.Nickname == "" {
		user.Nickname = httpUser.Username
	} else {
		user.Nickname = httpUser.Nickname
	}
	// 如果没有提供头像，则使用默认头像
	if httpUser.Avatar == "" {
		user.Avatar = model.DefaultAvatarURL
	} else {
		user.Avatar = httpUser.Avatar
	}
	user.Email = httpUser.Email

	// 执行其他字段的验证
	if errs := user.Validate(); len(errs) != 0 {
		klog.V(1).InfoS("validation failed", "username", httpUser.Username, "errors", errs.ToAggregate().Error())
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error()), nil)
		return
	}

	// 调用 service 创建用户
	if err := u.srv.Users().CreateUser(context.Background(), &user); err != nil {
		klog.V(1).InfoS("failed to create user", "username", httpUser.Username, "email", httpUser.Email, "error", err)
		// Store layer now decides the exact error code (e.g. ErrEmailAlreadyExist vs ErrUserAlreadyExist).
		core.WriteResponse(c, err, nil)
		return
	}

	if isAuthenticated {
		klog.V(1).InfoS("user created successfully by admin", "username", httpUser.Username, "role", user.Role, "status", user.Status, "requesterID", requesterIDAny)
	} else {
		klog.V(1).Info("user registered successfully")
	}
	core.WriteResponse(c, nil, nil)
}
