package local

import (
	"context"
	"os"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
)

// ServerConfigStore defines operations for server configuration files.
type ServerConfigStore interface {
	// Read reads the server configuration file content.
	Read(ctx context.Context) ([]byte, error)
	// Write writes the server configuration file.
	Write(ctx context.Context, content []byte) error
	// Delete deletes the server configuration file.
	Delete(ctx context.Context) error
}

type serverConfigStore struct{}

func newServerConfigStore() ServerConfigStore {
	return &serverConfigStore{}
}

func (s *serverConfigStore) Read(ctx context.Context) ([]byte, error) {
	serverConfPath := config.Get().WireGuard.ServerConfigPath()
	raw, err := os.ReadFile(serverConfPath)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConfPath, err)
	}
	return raw, nil
}

func (s *serverConfigStore) Write(ctx context.Context, content []byte) error {
	serverConfPath := config.Get().WireGuard.ServerConfigPath()

	// Preserve existing file perms where possible.
	var perm os.FileMode = 0600
	if st, statErr := os.Stat(serverConfPath); statErr == nil {
		perm = st.Mode().Perm()
	}

	if err := os.WriteFile(serverConfPath, content, perm); err != nil {
		return errors.WithCode(code.ErrWGWriteServerConfigFailed, "failed to write %s: %v", serverConfPath, err)
	}
	return nil
}

func (s *serverConfigStore) Delete(ctx context.Context) error {
	serverConfPath := config.Get().WireGuard.ServerConfigPath()
	if err := os.Remove(serverConfPath); err != nil && !os.IsNotExist(err) {
		return errors.WithCode(code.ErrWGServerConfigNotFound, "failed to delete %s: %v", serverConfPath, err)
	}
	return nil
}
