package wireguard

import (
	"encoding/base64"
	"os/exec"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// GeneratePrivateKey generates a new WireGuard private key using wg genkey command.
func GeneratePrivateKey() (string, error) {
	cmd := exec.Command("wg", "genkey")
	output, err := cmd.Output()
	if err != nil {
		klog.V(1).InfoS("failed to generate WireGuard private key", "error", err)
		return "", errors.WithCode(code.ErrWGKeyGenerationFailed, "failed to generate private key: %s", err.Error())
	}

	privateKey := strings.TrimSpace(string(output))
	if privateKey == "" {
		return "", errors.WithCode(code.ErrWGKeyGenerationFailed, "generated private key is empty")
	}

	// Validate base64 encoding
	_, err = base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "generated private key is not valid base64: %s", err.Error())
	}

	return privateKey, nil
}

// GeneratePublicKey generates a WireGuard public key from a private key using wg pubkey command.
func GeneratePublicKey(privateKey string) (string, error) {
	if privateKey == "" {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is empty")
	}

	// Validate base64 encoding
	_, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is not valid base64: %s", err.Error())
	}

	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	output, err := cmd.Output()
	if err != nil {
		klog.V(1).InfoS("failed to generate WireGuard public key", "error", err)
		return "", errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "failed to generate public key: %s", err.Error())
	}

	publicKey := strings.TrimSpace(string(output))
	if publicKey == "" {
		return "", errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "generated public key is empty")
	}

	return publicKey, nil
}

// GenerateKeyPair generates both private and public keys.
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privateKey, err = GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}

	publicKey, err = GeneratePublicKey(privateKey)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate public key from private key")
	}

	return privateKey, publicKey, nil
}

// ValidatePrivateKey validates that a string is a valid WireGuard private key.
func ValidatePrivateKey(privateKey string) error {
	if privateKey == "" {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is empty")
	}

	// Validate base64 encoding
	decoded, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is not valid base64: %s", err.Error())
	}

	// WireGuard private keys are 32 bytes
	if len(decoded) != 32 {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key must be 32 bytes, got %d", len(decoded))
	}

	return nil
}

// ValidatePublicKey validates that a string is a valid WireGuard public key.
func ValidatePublicKey(publicKey string) error {
	if publicKey == "" {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "public key is empty")
	}

	// Validate base64 encoding
	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "public key is not valid base64: %s", err.Error())
	}

	// WireGuard public keys are 32 bytes
	if len(decoded) != 32 {
		return errors.WithCode(code.ErrWGPrivateKeyInvalid, "public key must be 32 bytes, got %d", len(decoded))
	}

	return nil
}
