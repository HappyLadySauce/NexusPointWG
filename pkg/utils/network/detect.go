package network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

// DetectPublicIPByCommand detects the public IP address using system commands (curl/wget).
// This method is more reliable as it executes commands directly on the server.
// Returns the public IP address or an error if detection fails.
func DetectPublicIPByCommand() (string, error) {
	// List of commands to try (in order of preference)
	commands := []struct {
		name string
		args []string
	}{
		{"curl", []string{"-s", "--max-time", "3", "ifconfig.me"}},
		{"curl", []string{"-s", "--max-time", "3", "api.ipify.org"}},
		{"wget", []string{"-qO-", "--timeout=3", "ifconfig.me"}},
		{"wget", []string{"-qO-", "--timeout=3", "api.ipify.org"}},
	}

	for _, cmdSpec := range commands {
		cmd := exec.Command(cmdSpec.name, cmdSpec.args...)
		output, err := cmd.Output()
		if err != nil {
			klog.V(2).InfoS("failed to execute command", "command", cmdSpec.name, "error", err)
			continue
		}

		ipStr := strings.TrimSpace(string(output))
		ip := net.ParseIP(ipStr)
		if ip != nil && ip.To4() != nil {
			klog.V(2).InfoS("detected public IP using system command", "ip", ipStr, "command", cmdSpec.name)
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("failed to detect public IP using system commands")
}

// DetectPublicIP detects the public IP address using system commands (curl/wget).
// Returns the public IP address or an error if detection fails.
func DetectPublicIP(ctx context.Context) (string, error) {
	// Only use system commands (curl/wget) for detection
	publicIP, err := DetectPublicIPByCommand()
	if err == nil {
		return publicIP, nil
	}
	// System command detection failed, return error
	return "", fmt.Errorf("failed to detect public IP using system commands: %w", err)
}

// GetDefaultRouteInterfaceIP gets the IP address of the default route interface.
// Returns the IPv4 address of the interface used for default routing.
func GetDefaultRouteInterfaceIP() (string, error) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Try to find the default route interface by attempting to dial
	// We'll use a simple heuristic: find the interface that can reach a public IP
	// First, try to get the interface that would be used for a connection to 8.8.8.8
	var defaultInterface *net.Interface
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check if this interface has a non-loopback IPv4 address
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// Only consider IPv4 addresses
			if ip.To4() == nil {
				continue
			}

			// Skip loopback addresses
			if ip.IsLoopback() {
				continue
			}

			// Skip link-local addresses
			if ip.IsLinkLocalUnicast() {
				continue
			}

			// This looks like a valid interface, use it
			defaultInterface = &iface
			break
		}

		if defaultInterface != nil {
			break
		}
	}

	if defaultInterface == nil {
		return "", fmt.Errorf("no suitable network interface found")
	}

	// Get the first IPv4 address from the default interface
	addrs, err := defaultInterface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for interface %s: %w", defaultInterface.Name, err)
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}

		// Only return IPv4 addresses
		if ip.To4() != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found on interface %s", defaultInterface.Name)
}

// GetServerIP gets the server IP address, with automatic detection fallback.
// Priority: provided serverIP > system command detection > default route interface IP
func GetServerIP(ctx context.Context, serverIP string) (string, error) {
	// If serverIP is provided, use it
	if serverIP != "" {
		ip := net.ParseIP(serverIP)
		if ip != nil && ip.To4() != nil {
			return serverIP, nil
		}
		// Invalid IP provided, fall through to detection
	}

	// Try to detect public IP using system commands (curl/wget)
	publicIP, err := DetectPublicIP(ctx)
	if err == nil {
		// Verify it's a public IP (not private/local)
		ip := net.ParseIP(publicIP)
		if ip != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsPrivate() {
			klog.V(1).InfoS("detected server public IP", "ip", publicIP)
			return publicIP, nil
		}
		klog.V(1).InfoS("detected IP is not public, trying default route interface", "ip", publicIP)
	} else {
		klog.V(1).InfoS("failed to detect public IP using system commands, trying default route interface", "error", err)
	}

	// Fallback to default route interface IP
	defaultIP, err := GetDefaultRouteInterfaceIP()
	if err == nil {
		klog.V(1).InfoS("using default route interface IP", "ip", defaultIP)
		return defaultIP, nil
	}

	return "", fmt.Errorf("failed to detect server IP: public IP detection failed (%v), default route interface detection failed (%v)", err, err)
}
