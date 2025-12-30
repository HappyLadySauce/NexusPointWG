package wireguard

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/errors"
)

// GeneratePrivateKey generates a new WireGuard private key.
func GeneratePrivateKey(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "genkey")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrWGKeyGenerationFailed, "wg genkey failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// DerivePublicKey derives a public key from a private key.
func DerivePublicKey(ctx context.Context, privateKey string) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "pubkey")
	cmd.Stdin = bytes.NewBufferString(strings.TrimSpace(privateKey) + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "wg pubkey failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// ValidatePrivateKey validates a private key by attempting to derive its public key.
func ValidatePrivateKey(ctx context.Context, privateKey string) error {
	_, err := DerivePublicKey(ctx, privateKey)
	return err
}
