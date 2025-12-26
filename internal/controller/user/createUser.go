package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

// CreateUser create a user.
// @Summary Create user
// @Description Create a user with username, nickname, avatar, email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body v1.CreateUserRequest true "User information"
// @Success 200 {object} core.SuccessResponse "User created successfully"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input or validation failed"
// @Failure 401 {object} core.ErrResponse "Unauthorized - encryption error"
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
	user.Status = model.UserStatusActive
	// 注册只能创建普通用户，role 永远为 user
	user.Role = model.UserRoleUser

	// 执行其他字段的验证
	if errs := user.Validate(); len(errs) != 0 {
		klog.V(1).InfoS("validation failed", "username", httpUser.Username, "errors", errs.ToAggregate().Error())
		core.WriteResponse(c, errors.WithCode(code.ErrValidation, "%s", errs.ToAggregate().Error()), nil)
		return
	}

	// 调用 service 创建用户
	if err := u.srv.Users().CreateUser(context.Background(), &user); err != nil {
		klog.V(1).InfoS("failed to register user", "username", httpUser.Username, "email", httpUser.Email, "error", err)
		core.WriteResponse(c, err, nil)
		return
	}

	klog.V(1).Info("user registered successfully")
	core.WriteResponse(c, nil, nil)
}
