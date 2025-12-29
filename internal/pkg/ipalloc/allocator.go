package ipalloc

import (
	"fmt"
	"net/netip"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	iputil "github.com/HappyLadySauce/NexusPointWG/pkg/utils/ip"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// AllocationStats IP 分配统计信息
type AllocationStats struct {
	TotalHosts     int
	UsableHosts    int
	UsedCount      int
	AvailableCount int
}

// Allocator 提供 IP 地址分配功能
type Allocator struct {
	prefix   netip.Prefix
	serverIP netip.Addr
	used     map[netip.Addr]struct{}
}

// NewAllocator 创建一个新的 IP 分配器
func NewAllocator(prefix netip.Prefix, serverIP netip.Addr, usedIPs map[netip.Addr]struct{}) *Allocator {
	// 过滤掉不在前缀范围内的 IP
	used := make(map[netip.Addr]struct{})
	for ip := range usedIPs {
		if prefix.Contains(ip) {
			used[ip] = struct{}{}
		}
	}
	return &Allocator{
		prefix:   prefix,
		serverIP: serverIP,
		used:     used,
	}
}

// Stats 返回分配统计信息
func (a *Allocator) Stats() AllocationStats {
	totalHosts := 1 << (32 - a.prefix.Bits())
	usableHosts := totalHosts - 2 // 减去网络地址和广播地址
	if a.prefix.Bits() == 32 {
		usableHosts = 0
	}

	usedCount := len(a.used)
	if a.serverIP.IsValid() && a.prefix.Contains(a.serverIP) {
		start := a.prefix.Masked().Addr()
		last := iputil.LastIPv4(a.prefix)
		if a.serverIP != start && a.serverIP != last {
			// 服务器 IP 也算作已使用（如果不在 used map 中）
			if _, ok := a.used[a.serverIP]; !ok {
				usedCount++
			}
		}
	}

	availableCount := usableHosts - usedCount
	if availableCount < 0 {
		availableCount = 0
	}

	return AllocationStats{
		TotalHosts:     totalHosts,
		UsableHosts:    usableHosts,
		UsedCount:      usedCount,
		AvailableCount: availableCount,
	}
}

// Validate 验证 IP 是否可用于分配
func (a *Allocator) Validate(ip netip.Addr) error {
	if !ip.Is4() {
		return errors.WithCode(code.ErrIPNotIPv4, "IP is not IPv4: %s", ip)
	}

	if !a.prefix.Contains(ip) {
		return errors.WithCode(code.ErrIPOutOfRange, "IP %s is not within allocation prefix %s", ip, a.prefix)
	}

	start := a.prefix.Masked().Addr()
	last := iputil.LastIPv4(a.prefix)
	if ip == start {
		return errors.WithCode(code.ErrIPIsNetworkAddress, "IP %s is the network address", ip)
	}
	if ip == last {
		return errors.WithCode(code.ErrIPIsBroadcastAddress, "IP %s is the broadcast address", ip)
	}

	if ip == a.serverIP {
		return errors.WithCode(code.ErrIPIsServerIP, "IP %s is the server IP", ip)
	}

	if _, ok := a.used[ip]; ok {
		return errors.WithCode(code.ErrIPAlreadyInUse, "IP %s is already in use", ip)
	}

	return nil
}

// IsAvailable 检查 IP 是否可用
func (a *Allocator) IsAvailable(ip netip.Addr) bool {
	return a.Validate(ip) == nil
}

// Allocate 分配一个可用的 IP 地址
func (a *Allocator) Allocate() (netip.Addr, error) {
	start := a.prefix.Masked().Addr()
	last := iputil.LastIPv4(a.prefix)

	stats := a.Stats()
	klog.V(1).InfoS("allocating IP",
		"prefix", a.prefix,
		"total_hosts", stats.TotalHosts,
		"usable_hosts", stats.UsableHosts,
		"used", stats.UsedCount,
		"available", stats.AvailableCount)

	if stats.AvailableCount <= 0 {
		return netip.Addr{}, fmt.Errorf("no available IPs in %s (used: %d/%d)",
			a.prefix, stats.UsedCount, stats.UsableHosts)
	}

	// 从第一个可用主机开始迭代
	ip := start.Next()
	checked := 0
	for ip.Compare(last) < 0 {
		if !ip.Is4() {
			ip = ip.Next()
			continue
		}
		if !a.prefix.Contains(ip) {
			break
		}

		checked++
		if err := a.Validate(ip); err != nil {
			ip = ip.Next()
			continue
		}

		klog.V(1).InfoS("allocated IP", "ip", ip, "checked", checked)
		return ip, nil
	}

	return netip.Addr{}, fmt.Errorf("no available IP after checking %d addresses in %s (used: %d/%d)",
		checked, a.prefix, stats.UsedCount, stats.UsableHosts)
}
