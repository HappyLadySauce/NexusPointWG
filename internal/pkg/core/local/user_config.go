package local

import (
	"context"
	"os"
	"path/filepath"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
)

// UserConfigStore defines operations for user configuration files.
type UserConfigStore interface {
	// Read reads a user configuration file (e.g., peer.conf, privatekey, publickey, meta.json).
	Read(ctx context.Context, username, peerID, filename string) ([]byte, error)
	// Write writes a user configuration file.
	Write(ctx context.Context, username, peerID, filename string, content []byte) error
	// Delete deletes the user's configuration directory.
	Delete(ctx context.Context, username, peerID string) error
}

type userConfigStore struct{}

func newUserConfigStore() UserConfigStore {
	return &userConfigStore{}
}

func (u *userConfigStore) Read(ctx context.Context, username, peerID, filename string) ([]byte, error) {
	baseDir := filepath.Join(config.Get().WireGuard.ResolvedUserDir(), username, peerID)
	filePath := filepath.Join(baseDir, filename)
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGUserConfigNotFound, "failed to read %s: %v", filePath, err)
	}
	return b, nil
}

func (u *userConfigStore) Write(ctx context.Context, username, peerID, filename string, content []byte) error {
	baseDir := filepath.Join(config.Get().WireGuard.ResolvedUserDir(), username, peerID)
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return errors.WithCode(code.ErrWGUserDirCreateFailed, "failed to create user dir: %v", err)
	}

	filePath := filepath.Join(baseDir, filename)
	// Use appropriate file permissions based on file type
	var perm os.FileMode = 0600
	if err := os.WriteFile(filePath, content, perm); err != nil {
		return errors.WithCode(code.ErrWGConfigWriteFailed, "failed to write %s: %v", filePath, err)
	}
	return nil
}

func (u *userConfigStore) Delete(ctx context.Context, username, peerID string) error {
	baseDir := filepath.Join(config.Get().WireGuard.ResolvedUserDir(), username, peerID)
	if err := os.RemoveAll(baseDir); err != nil && !os.IsNotExist(err) {
		return errors.WithCode(code.ErrWGUserConfigNotFound, "failed to delete %s: %v", baseDir, err)
	}
	return nil
}
