package code

func init() {
	register(ErrUserAlreadyExist, 400, "User already exists")
	register(ErrEmailAlreadyExist, 400, "Email already exists")
	register(ErrUserNotFound, 404, "User not found")
	register(ErrUserNotActive, 403, "User account is not active")
	register(ErrSuccess, 200, "OK")
	register(ErrUnknown, 500, "Server error: Unknown server error")
	register(ErrBind, 400, "Error occurred while binding the request body to the struct")
	register(ErrValidation, 400, "Validation failed")
	register(ErrTokenInvalid, 401, "Token invalid")
	register(ErrDatabase, 500, "Server error: Database error")
	register(ErrEncrypt, 401, "Error occurred while encrypting the user password")
	register(ErrSignatureInvalid, 401, "Signature is invalid")
	register(ErrExpired, 401, "Token expired")
	register(ErrInvalidAuthHeader, 401, "Invalid authorization header")
	register(ErrMissingHeader, 401, "The `Authorization` header was empty")
	register(ErrPasswordIncorrect, 401, "Password was incorrect")
	register(ErrPermissionDenied, 403, "Permission denied")
	register(ErrEncodingFailed, 500, "Server error: Encoding failed due to an error with the data")
	register(ErrDecodingFailed, 500, "Server error: Decoding failed due to an error with the data")
	register(ErrInvalidJSON, 500, "Server error:Data is not valid JSON")
	register(ErrEncodingJSON, 500, "Server error: JSON data could not be encoded")
	register(ErrDecodingJSON, 500, "Server error: JSON data could not be decoded")
	register(ErrInvalidYaml, 500, "Server error:Data is not valid Yaml")
	register(ErrEncodingYaml, 500, "Server error: Yaml data could not be encoded")
	register(ErrDecodingYaml, 500, "Server error: Yaml data could not be decoded")
	register(ErrStoreNotInitialized, 500, "Server error: Store not initialized")

	// WireGuard: basic errors
	register(ErrWGPeerNotFound, 404, "WireGuard peer not found")
	register(ErrWGServerConfigNotFound, 500, "WireGuard server configuration file not found")
	register(ErrWGWriteServerConfigFailed, 500, "Failed to write WireGuard server configuration")
	register(ErrWGApplyFailed, 500, "Failed to apply WireGuard configuration")

	// WireGuard: IP address validation errors
	register(ErrIPNotIPv4, 400, "IP address is not IPv4")
	register(ErrIPOutOfRange, 400, "IP address is out of allocation prefix range")
	register(ErrIPIsNetworkAddress, 400, "IP address is a network address")
	register(ErrIPIsBroadcastAddress, 400, "IP address is a broadcast address")
	register(ErrIPIsServerIP, 400, "IP address is the server IP")
	register(ErrIPAlreadyInUse, 400, "IP address is already in use")

	// WireGuard: configuration errors
	register(ErrWGConfigNotInitialized, 500, "WireGuard configuration is not initialized")
	register(ErrWGLockAcquireFailed, 500, "Failed to acquire WireGuard lock")
	register(ErrWGServerPrivateKeyMissing, 500, "Server configuration missing Interface.PrivateKey")
	register(ErrWGServerAddressInvalid, 400, "Invalid server interface address")
	register(ErrWGAllowedIPsNotFound, 400, "AllowedIPs not found in server configuration")
	register(ErrWGIPv4PrefixNotFound, 400, "No valid IPv4 prefix found")
	register(ErrWGPrefixTooSmall, 400, "AllowedIPs prefix is too small to allocate client IP")
	register(ErrWGEndpointRequired, 400, "WireGuard endpoint is required")
	register(ErrWGIPAllocationFailed, 400, "IP address allocation failed")

	// WireGuard: key errors
	register(ErrWGPrivateKeyInvalid, 400, "Invalid WireGuard private key")
	register(ErrWGKeyGenerationFailed, 500, "Failed to generate WireGuard key")
	register(ErrWGPublicKeyGenerationFailed, 500, "Failed to generate public key from private key")

	// WireGuard: file operation errors
	register(ErrWGUserConfigNotFound, 404, "User WireGuard configuration not found")
	register(ErrWGPrivateKeyReadFailed, 500, "Failed to read private key file")
	register(ErrWGUserDirCreateFailed, 500, "Failed to create user directory")
	register(ErrWGPrivateKeyWriteFailed, 500, "Failed to write private key file")
	register(ErrWGPublicKeyWriteFailed, 500, "Failed to write public key file")
	register(ErrWGConfigWriteFailed, 500, "Failed to write WireGuard configuration file")

	// WireGuard: data errors
	register(ErrWGPeerIDGenerationFailed, 500, "Failed to generate peer ID")
	register(ErrWGPeerNil, 400, "Peer is nil")

	// WireGuard: IP pool errors
	register(ErrIPPoolNotFound, 404, "IP pool not found")
	register(ErrIPPoolAlreadyExists, 400, "IP pool with the same CIDR already exists")
	register(ErrIPPoolInvalidCIDR, 400, "Invalid CIDR format for IP pool")
	register(ErrIPPoolInUse, 400, "IP pool is in use and cannot be deleted")
	register(ErrIPPoolDisabled, 400, "IP pool is disabled")
}
