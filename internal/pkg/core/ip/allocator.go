package ip

import (
	"context"
	"net"
	"sort"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/errors"
)

// Allocator handles IP address allocation for WireGuard peers.
type Allocator struct {
	store store.Factory
}

// NewAllocator creates a new IP allocator.
func NewAllocator(store store.Factory) *Allocator {
	return &Allocator{store: store}
}

// AllocateIP allocates an IP address from the specified IP pool.
// If preferredIP is provided and available, it will be used; otherwise, the first available IP will be allocated.
func (a *Allocator) AllocateIP(ctx context.Context, poolID, preferredIP string) (string, error) {
	// Get IP pool
	pool, err := a.store.IPPools().GetIPPool(ctx, poolID)
	if err != nil {
		return "", err
	}

	if pool.Status != model.IPPoolStatusActive {
		return "", errors.WithCode(code.ErrIPPoolDisabled, "IP pool %s is disabled", poolID)
	}

	// Get all allocated IPs for this pool
	allocatedIPs, err := a.store.IPAllocations().GetAllocatedIPsByPoolID(ctx, poolID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get allocated IPs")
	}

	// Convert allocated IPs to a map for quick lookup
	allocatedMap := make(map[string]bool)
	for _, ip := range allocatedIPs {
		allocatedMap[ip] = true
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(pool.CIDR)
	if err != nil {
		return "", errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid CIDR format: %s", pool.CIDR)
	}

	// Extract server IP
	serverIPStr, _ := ExtractIPFromCIDR(pool.ServerIP)

	// If preferred IP is provided, validate and use it
	if preferredIP != "" {
		// Validate preferred IP
		if err := ValidateIPv4(preferredIP); err != nil {
			return "", err
		}
		if err := ValidateIPInCIDR(preferredIP, pool.CIDR); err != nil {
			return "", err
		}
		if err := ValidateIPNotReserved(preferredIP, pool.CIDR, serverIPStr); err != nil {
			return "", err
		}

		// Check if already allocated
		if allocatedMap[preferredIP] {
			return "", errors.WithCode(code.ErrIPAlreadyInUse, "IP address %s is already in use", preferredIP)
		}

		return preferredIP, nil
	}

	// Auto-allocate: find the first available IP
	availableIP, err := a.findFirstAvailableIP(ipNet, serverIPStr, allocatedMap)
	if err != nil {
		return "", errors.WithCode(code.ErrWGIPAllocationFailed, "failed to allocate IP address: %s", err.Error())
	}

	return availableIP, nil
}

// findFirstAvailableIP finds the first available IP address in the network range.
func (a *Allocator) findFirstAvailableIP(ipNet *net.IPNet, serverIPStr string, allocatedMap map[string]bool) (string, error) {
	networkIP := ipNet.IP
	mask := ipNet.Mask

	// Calculate network size
	ones, bits := mask.Size()
	if bits != 32 {
		return "", errors.WithCode(code.ErrIPNotIPv4, "only IPv4 networks are supported")
	}

	// Generate all possible IPs in the network
	var candidateIPs []string
	for i := 1; i < (1<<(32-ones))-1; i++ {
		ip := make(net.IP, 4)
		copy(ip, networkIP)

		// Add offset to network IP
		offset := i
		for j := 3; j >= 0 && offset > 0; j-- {
			ip[j] += byte(offset & 0xff)
			offset >>= 8
		}

		ipStr := ip.String()

		// Skip network address, broadcast address, and server IP
		if ip.Equal(networkIP) {
			continue
		}

		// Check broadcast address
		broadcastIP := make(net.IP, len(networkIP))
		copy(broadcastIP, networkIP)
		for j := range broadcastIP {
			broadcastIP[j] |= ^mask[j]
		}
		if ip.Equal(broadcastIP) {
			continue
		}

		// Check server IP
		if serverIPStr != "" {
			serverIP := net.ParseIP(serverIPStr)
			if serverIP != nil && ip.Equal(serverIP) {
				continue
			}
		}

		// Check if already allocated
		if !allocatedMap[ipStr] {
			candidateIPs = append(candidateIPs, ipStr)
		}
	}

	if len(candidateIPs) == 0 {
		return "", errors.WithCode(code.ErrWGIPAllocationFailed, "no available IP addresses in pool")
	}

	// Sort IPs to ensure consistent allocation order
	sort.Strings(candidateIPs)

	return candidateIPs[0], nil
}

// ValidateAndAllocateIP validates an IP address and allocates it if valid.
// This is used when an IP is manually specified.
func (a *Allocator) ValidateAndAllocateIP(ctx context.Context, poolID, ipStr string) error {
	// Get IP pool
	pool, err := a.store.IPPools().GetIPPool(ctx, poolID)
	if err != nil {
		return err
	}

	if pool.Status != model.IPPoolStatusActive {
		return errors.WithCode(code.ErrIPPoolDisabled, "IP pool %s is disabled", poolID)
	}

	// Extract server IP
	serverIPStr, _ := ExtractIPFromCIDR(pool.ServerIP)

	// Validate IP
	if err := ValidateIPv4(ipStr); err != nil {
		return err
	}
	if err := ValidateIPInCIDR(ipStr, pool.CIDR); err != nil {
		return err
	}
	if err := ValidateIPNotReserved(ipStr, pool.CIDR, serverIPStr); err != nil {
		return err
	}

	// Check if already allocated
	allocation, err := a.store.IPAllocations().GetIPAllocationByIPAddress(ctx, ipStr)
	if err == nil && allocation != nil && allocation.Status == model.IPAllocationStatusAllocated {
		return errors.WithCode(code.ErrIPAlreadyInUse, "IP address %s is already in use", ipStr)
	}

	return nil
}

// GetAvailableIPs returns a list of available IP addresses in the pool.
func (a *Allocator) GetAvailableIPs(ctx context.Context, poolID string, limit int) ([]string, error) {
	// Get IP pool
	pool, err := a.store.IPPools().GetIPPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	if pool.Status != model.IPPoolStatusActive {
		return nil, errors.WithCode(code.ErrIPPoolDisabled, "IP pool %s is disabled", poolID)
	}

	// Get all allocated IPs for this pool
	allocatedIPs, err := a.store.IPAllocations().GetAllocatedIPsByPoolID(ctx, poolID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get allocated IPs")
	}

	// Convert allocated IPs to a map for quick lookup
	allocatedMap := make(map[string]bool)
	for _, ip := range allocatedIPs {
		allocatedMap[ip] = true
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(pool.CIDR)
	if err != nil {
		return nil, errors.WithCode(code.ErrIPPoolInvalidCIDR, "invalid CIDR format: %s", pool.CIDR)
	}

	// Extract server IP
	serverIPStr, _ := ExtractIPFromCIDR(pool.ServerIP)

	// Generate available IPs
	var availableIPs []string
	networkIP := ipNet.IP
	mask := ipNet.Mask

	ones, bits := mask.Size()
	if bits != 32 {
		return nil, errors.WithCode(code.ErrIPNotIPv4, "only IPv4 networks are supported")
	}

	for i := 1; i < (1<<(32-ones))-1 && len(availableIPs) < limit; i++ {
		ip := make(net.IP, 4)
		copy(ip, networkIP)

		// Add offset to network IP
		offset := i
		for j := 3; j >= 0 && offset > 0; j-- {
			ip[j] += byte(offset & 0xff)
			offset >>= 8
		}

		ipStr := ip.String()

		// Skip network address, broadcast address, and server IP
		if ip.Equal(networkIP) {
			continue
		}

		broadcastIP := make(net.IP, len(networkIP))
		copy(broadcastIP, networkIP)
		for j := range broadcastIP {
			broadcastIP[j] |= ^mask[j]
		}
		if ip.Equal(broadcastIP) {
			continue
		}

		if serverIPStr != "" {
			serverIP := net.ParseIP(serverIPStr)
			if serverIP != nil && ip.Equal(serverIP) {
				continue
			}
		}

		// Check if already allocated
		if !allocatedMap[ipStr] {
			availableIPs = append(availableIPs, ipStr)
		}
	}

	// Sort IPs
	sort.Strings(availableIPs)

	return availableIPs, nil
}

// ReleaseIP releases an allocated IP address.
func (a *Allocator) ReleaseIP(ctx context.Context, peerID string) error {
	allocation, err := a.store.IPAllocations().GetIPAllocationByPeerID(ctx, peerID)
	if err != nil {
		return errors.Wrap(err, "failed to get IP allocation")
	}

	if allocation == nil {
		return nil // Already released or never allocated
	}

	allocation.Status = model.IPAllocationStatusReleased
	if err := a.store.IPAllocations().UpdateIPAllocation(ctx, allocation); err != nil {
		return errors.Wrap(err, "failed to release IP allocation")
	}

	return nil
}

