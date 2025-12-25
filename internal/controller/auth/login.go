package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/core"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/jwt"
	"github.com/HappyLadySauce/errors"
)

// Login handles user login request.
// @Summary User login
// @Description Authenticate user with username and password, returns JWT token string
// @Tags auth
// @Accept json
// @Produce json
// @Param login body v1.LoginRequest true "Login credentials"
// @Success 200 {object} v1.LoginResponse "Login successful"
// @Failure 400 {object} core.ErrResponse "Bad request - invalid input"
// @Failure 401 {object} core.ErrResponse "Unauthorized - invalid credentials"
// @Failure 403 {object} core.ErrResponse "Forbidden - user account is not active"
// @Router /api/v1/login [post]
func (a *AuthController) Login(c *gin.Context) {
	klog.V(1).Info("login function called.")

	var loginReq v1.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		klog.Errorf("invalid request body: %v", err)
		core.WriteResponseBindErr(c, err, nil)
		return
	}

	// 验证用户名和密码
	user, err := a.srv.Auth().Login(context.Background(), loginReq.Username, loginReq.Password)
	if err != nil {
		klog.Errorf("login failed: %v", err)
		core.WriteResponse(c, err, nil)
		return
	}

	// 生成 JWT token
	cfg := config.Get()
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, cfg.JWT.Secret, cfg.JWT.Expiration)
	if err != nil {
		klog.Errorf("failed to generate token: %v", err)
		core.WriteResponse(c, errors.WithCode(code.ErrEncrypt, err.Error()), nil)
		return
	}

	response := v1.LoginResponse{
		Token: token,
	}

	klog.V(1).Info("login successful")
	core.WriteResponse(c, nil, response)
}
