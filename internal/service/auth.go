package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/passwd"
	"github.com/HappyLadySauce/errors"
)

type AuthSrv interface {
	Login(ctx context.Context, username, password string) (*model.User, error)
}

type authSrv struct {
	store store.Factory
}

// AuthSrv if implemented, then authSrv implements AuthSrv interface.
var _ AuthSrv = (*authSrv)(nil)

func newAuth(s *service) *authSrv {
	return &authSrv{store: s.store}
}

func (a *authSrv) Login(ctx context.Context, username, password string) (*model.User, error) {
	// 根据用户名获取用户
	user, err := a.store.Users().GetUserByUsername(ctx, username)
	if err != nil {
		// 用户不存在或数据库错误
		return nil, errors.WithCode(code.ErrPasswordIncorrect, "invalid username or password")
	}

	// 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, errors.WithCode(code.ErrPermissionDenied, "user account is not active")
	}

	// 验证密码
	if !passwd.VerifyPassword(password, user.Salt, user.PasswordHash) {
		return nil, errors.WithCode(code.ErrPasswordIncorrect, "invalid username or password")
	}

	return user, nil
}
