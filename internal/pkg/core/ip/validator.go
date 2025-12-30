package ip

import (
	"fmt"
	"net"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/errors"
)

// ValidateIPv4 validates that the given IP address is a valid IPv4 address.
func ValidateIPv4(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "invalid IP address format: %s", ipStr)
	}
	if ip.To4() == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "IP address is not IPv4: %s", ipStr)
	}
	return nil
}

// ValidateIPInCIDR validates that the given IP address is within the CIDR range.
func ValidateIPInCIDR(ipStr, cidrStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "invalid IP address format: %s", ipStr)
	}
	if ip.To4() == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "IP address is not IPv4: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid CIDR format: %s", cidrStr)
	}

	if !ipNet.Contains(ip) {
		return errors.WithCode(code.ErrIPOutOfRange, "IP address %s is not in CIDR range %s", ipStr, cidrStr)
	}

	return nil
}

// ValidateIPNotReserved validates that the IP address is not a reserved address.
// Reserved addresses include: network address, broadcast address, and server IP.
func ValidateIPNotReserved(ipStr, cidrStr, serverIPStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "invalid IP address format: %s", ipStr)
	}
	if ip.To4() == nil {
		return errors.WithCode(code.ErrIPNotIPv4, "IP address is not IPv4: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid CIDR format: %s", cidrStr)
	}

	// Check if IP is network address (first IP in the network)
	networkIP := ipNet.IP
	if ip.Equal(networkIP) {
		return errors.WithCode(code.ErrIPIsNetworkAddress, "IP address %s is the network address", ipStr)
	}

	// Check if IP is broadcast address (last IP in the network)
	broadcastIP := make(net.IP, len(networkIP))
	copy(broadcastIP, networkIP)
	for i := range broadcastIP {
		broadcastIP[i] |= ^ipNet.Mask[i]
	}
	if ip.Equal(broadcastIP) {
		return errors.WithCode(code.ErrIPIsBroadcastAddress, "IP address %s is the broadcast address", ipStr)
	}

	// Check if IP is server IP
	if serverIPStr != "" {
		serverIP := net.ParseIP(serverIPStr)
		if serverIP != nil && ip.Equal(serverIP) {
			return errors.WithCode(code.ErrIPIsServerIP, "IP address %s is the server IP", ipStr)
		}
	}

	return nil
}

// ExtractIPFromCIDR extracts the IP address from a CIDR string (e.g., "100.100.100.2/32" -> "100.100.100.2").
func ExtractIPFromCIDR(cidrStr string) (string, error) {
	ipStr, _, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return "", errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid CIDR format: %s", cidrStr)
	}
	return ipStr.String(), nil
}

// FormatIPAsCIDR formats an IP address as a CIDR with /32 prefix (e.g., "100.100.100.2" -> "100.100.100.2/32").
func FormatIPAsCIDR(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", errors.WithCode(code.ErrIPNotIPv4, "invalid IP address format: %s", ipStr)
	}
	if ip.To4() == nil {
		return "", errors.WithCode(code.ErrIPNotIPv4, "IP address is not IPv4: %s", ipStr)
	}
	return fmt.Sprintf("%s/32", ipStr), nil
}

// ExtractIPFromEndpoint extracts the IP address from an endpoint string (e.g., "118.24.41.142:51820" -> "118.24.41.142").
func ExtractIPFromEndpoint(endpoint string) (string, error) {
	if endpoint == "" {
		return "", nil
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		// If SplitHostPort fails, try parsing as IP only
		if net.ParseIP(endpoint) != nil {
			return endpoint, nil
		}
		return "", errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid endpoint format: %s", endpoint)
	}
	return host, nil
}
