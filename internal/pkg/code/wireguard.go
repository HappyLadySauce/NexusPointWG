package code

// WireGuard: wireguard-related errors.
// Code must start with 1xxxxx.
const (
	// ErrWGPeerNotFound - 404: WireGuard peer not found.
	ErrWGPeerNotFound int = iota + 120001

	// ErrWGServerConfigNotFound - 500: Server WireGuard config not found or unreadable.
	ErrWGServerConfigNotFound

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
)
