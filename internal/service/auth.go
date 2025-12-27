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
		// SECURITY: prevent user enumeration attacks.
		// If the user does not exist, return the same error as an incorrect password.
		// This prevents attackers from distinguishing between "user doesn't exist" and "wrong password"
		// responses, which would allow them to enumerate valid usernames in the system.
		if errors.ParseCoder(err).Code() == code.ErrUserNotFound {
			return nil, errors.WithCode(code.ErrPasswordIncorrect, "%s", code.Message(code.ErrPasswordIncorrect))
		}

		// Other errors are treated as internal server/database errors.
		return nil, errors.WithCode(code.ErrDatabase, "%s", code.Message(code.ErrDatabase))
	}

	// 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, errors.WithCode(code.ErrUserNotActive, "%s", code.Message(code.ErrUserNotActive))
	}

	// 验证密码
	if !passwd.VerifyPassword(password, user.Salt, user.PasswordHash) {
		return nil, errors.WithCode(code.ErrPasswordIncorrect, "%s", code.Message(code.ErrPasswordIncorrect))
	}

	return user, nil
}
