package sqlite

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/errors"
)

type users struct {
	db *gorm.DB
}

func newUsers(ds *datastore) *users {
	return &users{ds.db}
}

func (u *users) GetUser(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrDatabase, err.Error())
		}
		return nil, err
	}
	return &user, nil
}

func (u *users) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrDatabase, err.Error())
		}
		return nil, err
	}
	return &user, nil
}

func (u *users) CreateUser(ctx context.Context, user *model.User) error {
	err := u.db.WithContext(ctx).Create(user).Error
	if err != nil {
		// Check if it's a unique constraint violation (e.g., duplicate username)
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrUserAlreadyExist, err.Error())
		}
		return errors.WithCode(code.ErrDatabase, err.Error())
	}
	return nil
}

func (u *users) UpdateUser(ctx context.Context, user *model.User) error {
	err := u.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return errors.WithCode(code.ErrDatabase, err.Error())
	}
	return nil
}

func (u *users) DeleteUser(ctx context.Context, id string) error {
	err := u.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.WithCode(code.ErrDatabase, err.Error())
	}
	return nil
}

// isUniqueConstraintError checks if the error is a unique constraint violation.
// SQLite returns various error messages for unique constraint violations:
// - "UNIQUE constraint failed: ..."
// - "Duplicate entry '...' for key '...'"
// - Error code 2067 (SQLITE_CONSTRAINT_UNIQUE)
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Check for common unique constraint error patterns
	uniquePatterns := []string{
		"unique constraint failed",
		"duplicate entry",
		"constraint failed",
		"sqlite_constraint_unique",
	}

	for _, pattern := range uniquePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
