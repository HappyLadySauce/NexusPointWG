package code

// WireGuard: wireguard-related errors.
// Code must start with 1xxxxx.
const (
	// ErrWGPeerNotFound - 404: WireGuard peer not found.
	ErrWGPeerNotFound int = iota + 120001

	// ErrWGServerConfigNotFound - 500: Server WireGuard config not found or unreadable.
	ErrWGServerConfigNotFound

	// ErrWGServerConfigParseFailed - 500: Failed to parse server WireGuard config.
	ErrWGServerConfigParseFailed

	// ErrWGWriteServerConfigFailed - 500: Failed to write server WireGuard config.
	ErrWGWriteServerConfigFailed

	// ErrWGApplyFailed - 500: Failed to apply WireGuard config via systemd.
	ErrWGApplyFailed

	// ErrIPNotIPv4 - 400: IP is not IPv4.
	ErrIPNotIPv4

	// ErrIPOutOfRange - 400: IP is not within allocation prefix.
	ErrIPOutOfRange

	// ErrIPIsNetworkAddress - 400: IP is the network address.
	ErrIPIsNetworkAddress

	// ErrIPIsBroadcastAddress - 400: IP is the broadcast address.
	ErrIPIsBroadcastAddress

	// ErrIPIsServerIP - 400: IP is the server IP.
	ErrIPIsServerIP

	// ErrIPAlreadyInUse - 400: IP is already in use.
	ErrIPAlreadyInUse

	// WireGuard 配置相关错误
	// ErrWGConfigNotInitialized - 500: WireGuard config is not initialized.
	ErrWGConfigNotInitialized int = iota + 120010

	// ErrWGLockAcquireFailed - 500: Failed to acquire WireGuard lock.
	ErrWGLockAcquireFailed

	// ErrWGServerPrivateKeyMissing - 500: Server config missing Interface.PrivateKey.
	ErrWGServerPrivateKeyMissing

	// ErrWGServerAddressInvalid - 400: Invalid server interface address.
	ErrWGServerAddressInvalid

	// ErrWGAllowedIPsNotFound - 400: No AllowedIPs found in server config.
	ErrWGAllowedIPsNotFound

	// ErrWGIPv4PrefixNotFound - 400: No valid IPv4 prefix found in AllowedIPs.
	ErrWGIPv4PrefixNotFound

	// ErrWGPrefixTooSmall - 400: AllowedIPs prefix too small for client IP allocation.
	ErrWGPrefixTooSmall

	// ErrWGEndpointRequired - 400: WireGuard endpoint is required.
	ErrWGEndpointRequired

	// ErrWGIPAllocationFailed - 400: Failed to allocate IP address.
	ErrWGIPAllocationFailed

	// WireGuard 密钥相关错误
	// ErrWGPrivateKeyInvalid - 400: Invalid private key.
	ErrWGPrivateKeyInvalid int = iota + 120020

	// ErrWGKeyGenerationFailed - 500: Failed to generate WireGuard key.
	ErrWGKeyGenerationFailed

	// ErrWGPublicKeyGenerationFailed - 500: Failed to generate public key from private key.
	ErrWGPublicKeyGenerationFailed

	// WireGuard 文件操作错误
	// ErrWGUserConfigNotFound - 404: User WireGuard config not found.
	ErrWGUserConfigNotFound int = iota + 120030

	// ErrWGPrivateKeyReadFailed - 500: Failed to read private key file.
	ErrWGPrivateKeyReadFailed

	// ErrWGUserDirCreateFailed - 500: Failed to create user directory.
	ErrWGUserDirCreateFailed

	// ErrWGPrivateKeyWriteFailed - 500: Failed to write private key file.
	ErrWGPrivateKeyWriteFailed

	// ErrWGPublicKeyWriteFailed - 500: Failed to write public key file.
	ErrWGPublicKeyWriteFailed

	// ErrWGConfigWriteFailed - 500: Failed to write WireGuard config file.
	ErrWGConfigWriteFailed

	// WireGuard 数据相关错误
	// ErrWGPeerIDGenerationFailed - 500: Failed to generate peer ID.
	ErrWGPeerIDGenerationFailed int = iota + 120040

	// ErrWGPeerNil - 400: Peer is nil.
	ErrWGPeerNil
)
