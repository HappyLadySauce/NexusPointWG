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
		// Check if it's a unique constraint violation (e.g., duplicate username or email)
		if isUniqueConstraintError(err) {
			// Parse the error message to determine which field caused the constraint violation
			field := parseUniqueConstraintField(err)
			// Decide which coded error to return based on the duplicated field.
			// - username duplicate -> ErrUserAlreadyExist
			// - email duplicate -> ErrEmailAlreadyExist
			if field == "email" {
				return errors.WithCode(code.ErrEmailAlreadyExist, "%s", code.Message(code.ErrEmailAlreadyExist))
			}
			// Default to username duplicate / generic already-exists
			return errors.WithCode(code.ErrUserAlreadyExist, "%s", code.Message(code.ErrUserAlreadyExist))
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
			field := parseUniqueConstraintField(err)
			if field == "email" {
				return errors.WithCode(code.ErrEmailAlreadyExist, "%s", code.Message(code.ErrEmailAlreadyExist))
			}
			return errors.WithCode(code.ErrUserAlreadyExist, "%s", code.Message(code.ErrUserAlreadyExist))
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

// parseUniqueConstraintField extracts the field name from a unique constraint error.
// Returns "username", "email", or empty string if the field cannot be determined.
func parseUniqueConstraintField(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	// Check for username constraint
	if strings.Contains(errMsgLower, "users.username") ||
		strings.Contains(errMsgLower, ".username") ||
		strings.Contains(errMsgLower, ": username") {
		return "username"
	}

	// Check for email constraint
	if strings.Contains(errMsgLower, "users.email") ||
		strings.Contains(errMsgLower, ".email") ||
		strings.Contains(errMsgLower, ": email") {
		return "email"
	}

	return ""
}
