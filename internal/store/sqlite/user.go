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
			return nil, errors.WithCode(code.ErrUserNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &user, nil
}

func (u *users) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &user, nil
}

func (u *users) CreateUser(ctx context.Context, user *model.User) error {
	err := u.db.WithContext(ctx).Create(user).Error
	if err != nil {
		// Check if it's a unique constraint violation (e.g., duplicate username)
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrUserAlreadyExist, "%s", err.Error())
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (u *users) UpdateUser(ctx context.Context, user *model.User) error {
	err := u.db.WithContext(ctx).Save(user).Error
	if err != nil {
		// Check if it's a unique constraint violation (e.g., duplicate username or email)
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrUserAlreadyExist, "%s", err.Error())
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (u *users) DeleteUser(ctx context.Context, id string) error {
	err := u.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return nil to allow idempotent delete operations
			return nil
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

// isUniqueConstraintError checks if the error is a unique constraint violation.
// SQLite (via GORM) returns various error messages for unique constraint violations:
// - "UNIQUE constraint failed: users.username" (SQLite native format)
// - "UNIQUE constraint failed: users.email" (SQLite native format)
// - "constraint failed: UNIQUE constraint failed: users.username" (wrapped format)
// - "sqlite: UNIQUE constraint failed: users.username" (with driver prefix)
//
// Note: The error code 2067 (SQLITE_CONSTRAINT_UNIQUE) is typically not exposed
// directly by GORM, so we rely on error message pattern matching.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Check for SQLite unique constraint error patterns
	// Order matters: more specific patterns first
	uniquePatterns := []string{
		"unique constraint failed", // Most common SQLite format
		"sqlite_constraint_unique", // SQLite error code name
		"constraint failed",        // Generic constraint failure (may match other constraints, but unique is most common)
	}

	for _, pattern := range uniquePatterns {
		if strings.Contains(errMsg, pattern) {
			// Additional check: ensure it's not a different constraint type
			// If the error mentions "NOT NULL" or "FOREIGN KEY", it's not a unique constraint
			if strings.Contains(errMsg, "not null") || strings.Contains(errMsg, "foreign key") {
				continue
			}
			return true
		}
	}

	return false
}
