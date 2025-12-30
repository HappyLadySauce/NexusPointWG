package code

// WireGuard: basic errors (120001-120004)
const (
	// ErrWGPeerNotFound - 404: WireGuard peer not found.
	ErrWGPeerNotFound int = iota + 120001

	// ErrWGServerConfigNotFound - 500: WireGuard server configuration file not found.
	ErrWGServerConfigNotFound

	// ErrWGWriteServerConfigFailed - 500: Failed to write WireGuard server configuration.
	ErrWGWriteServerConfigFailed

	// ErrWGApplyFailed - 500: Failed to apply WireGuard configuration.
	ErrWGApplyFailed
)

// WireGuard: IP address validation errors (120005-120010)
const (
	// ErrIPNotIPv4 - 400: IP address is not IPv4.
	ErrIPNotIPv4 int = iota + 120005

	// ErrIPOutOfRange - 400: IP address is out of allocation prefix range.
	ErrIPOutOfRange

	// ErrIPIsNetworkAddress - 400: IP address is a network address.
	ErrIPIsNetworkAddress

	// ErrIPIsBroadcastAddress - 400: IP address is a broadcast address.
	ErrIPIsBroadcastAddress

	// ErrIPIsServerIP - 400: IP address is the server IP.
	ErrIPIsServerIP

	// ErrIPAlreadyInUse - 400: IP address is already in use.
	ErrIPAlreadyInUse
)

// WireGuard: configuration errors (120011-120019)
const (
	// ErrWGConfigNotInitialized - 500: WireGuard configuration is not initialized.
	ErrWGConfigNotInitialized int = iota + 120011

	// ErrWGLockAcquireFailed - 500: Failed to acquire WireGuard lock.
	ErrWGLockAcquireFailed

	// ErrWGServerPrivateKeyMissing - 500: Server configuration missing Interface.PrivateKey.
	ErrWGServerPrivateKeyMissing

	// ErrWGServerAddressInvalid - 400: Invalid server interface address.
	ErrWGServerAddressInvalid

	// ErrWGAllowedIPsNotFound - 400: AllowedIPs not found in server configuration.
	ErrWGAllowedIPsNotFound

	// ErrWGIPv4PrefixNotFound - 400: No valid IPv4 prefix found.
	ErrWGIPv4PrefixNotFound

	// ErrWGPrefixTooSmall - 400: AllowedIPs prefix is too small to allocate client IP.
	ErrWGPrefixTooSmall

	// ErrWGEndpointRequired - 400: WireGuard endpoint is required.
	ErrWGEndpointRequired

	// ErrWGIPAllocationFailed - 400: IP address allocation failed.
	ErrWGIPAllocationFailed
)

// WireGuard: key errors (120020-120022)
const (
	// ErrWGPrivateKeyInvalid - 400: Invalid WireGuard private key.
	ErrWGPrivateKeyInvalid int = iota + 120020

	// ErrWGKeyGenerationFailed - 500: Failed to generate WireGuard key.
	ErrWGKeyGenerationFailed

	// ErrWGPublicKeyGenerationFailed - 500: Failed to generate public key from private key.
	ErrWGPublicKeyGenerationFailed
)

// WireGuard: file operation errors (120030-120035)
const (
	// ErrWGUserConfigNotFound - 404: User WireGuard configuration not found.
	ErrWGUserConfigNotFound int = iota + 120030

	// ErrWGPrivateKeyReadFailed - 500: Failed to read private key file.
	ErrWGPrivateKeyReadFailed

	// ErrWGUserDirCreateFailed - 500: Failed to create user directory.
	ErrWGUserDirCreateFailed

	// ErrWGPrivateKeyWriteFailed - 500: Failed to write private key file.
	ErrWGPrivateKeyWriteFailed

	// ErrWGPublicKeyWriteFailed - 500: Failed to write public key file.
	ErrWGPublicKeyWriteFailed

	// ErrWGConfigWriteFailed - 500: Failed to write WireGuard configuration file.
	ErrWGConfigWriteFailed
)

// WireGuard: data errors (120040-120041)
const (
	// ErrWGPeerIDGenerationFailed - 500: Failed to generate peer ID.
	ErrWGPeerIDGenerationFailed int = iota + 120040

	// ErrWGPeerNil - 400: Peer is nil.
	ErrWGPeerNil
)

// WireGuard: IP pool errors (120050-120055)
const (
	// ErrIPPoolNotFound - 404: IP pool not found.
	ErrIPPoolNotFound int = iota + 120050

	// ErrIPPoolAlreadyExists - 400: IP pool with the same CIDR already exists.
	ErrIPPoolAlreadyExists

	// ErrIPPoolInvalidCIDR - 400: Invalid CIDR format for IP pool.
	ErrIPPoolInvalidCIDR

	// ErrIPPoolInUse - 400: IP pool is in use and cannot be deleted.
	ErrIPPoolInUse

	// ErrIPPoolDisabled - 400: IP pool is disabled.
	ErrIPPoolDisabled
)
