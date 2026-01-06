package wireguard

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/errors"
	"golang.org/x/crypto/curve25519"
	"k8s.io/klog/v2"
)

// GeneratePrivateKey generates a new WireGuard private key.
// Uses pure Go implementation (crypto/rand + Curve25519) instead of wg command.
// This ensures it works in Docker containers without wireguard-tools installed.
func GeneratePrivateKey() (string, error) {
	// Try pure Go implementation first (works everywhere)
	privateKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privateKeyBytes); err != nil {
		klog.V(1).InfoS("failed to generate random bytes for private key", "error", err)
		return "", errors.WithCode(code.ErrWGKeyGenerationFailed, "failed to generate random bytes: %s", err.Error())
	}

	// WireGuard private key clamping (required for Curve25519)
	// Set the least significant bit of the first byte to 0
	privateKeyBytes[0] &= 248
	// Set the second least significant bit of the last byte to 0
	privateKeyBytes[31] &= 127
	// Set the third least significant bit of the last byte to 1
	privateKeyBytes[31] |= 64

	// Encode to base64
	privateKey := base64.StdEncoding.EncodeToString(privateKeyBytes)
	return privateKey, nil
}

// GeneratePublicKey generates a WireGuard public key from a private key.
// Uses pure Go implementation (Curve25519) instead of wg command.
// This ensures it works in Docker containers without wireguard-tools installed.
func GeneratePublicKey(privateKey string) (string, error) {
	if privateKey == "" {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is empty")
	}

	// Decode private key from base64
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key is not valid base64: %s", err.Error())
	}

	// Validate length
	if len(privateKeyBytes) != 32 {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "private key must be 32 bytes, got %d", len(privateKeyBytes))
	}

	// Generate public key using Curve25519
	var publicKeyBytes [32]byte
	var privateKeyArray [32]byte
	copy(privateKeyArray[:], privateKeyBytes)
	curve25519.ScalarBaseMult(&publicKeyBytes, &privateKeyArray)

	// Encode to base64
	publicKey := base64.StdEncoding.EncodeToString(publicKeyBytes[:])
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
