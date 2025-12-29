package ipalloc

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"k8s.io/klog/v2"
)

// ParseFirstV4PrefixFromAllowedIPs 从 AllowedIPs 中解析第一个 IPv4 前缀
// AllowedIPs is comma-separated, e.g., "100.100.100.0/24, 192.168.1.0/24"
func ParseFirstV4PrefixFromAllowedIPs(allowedIPsLine string) (netip.Prefix, error) {
	parts := strings.Split(allowedIPsLine, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(p)
		if err != nil {
			continue
		}
		if prefix.Addr().Is4() {
			return prefix.Masked(), nil
		}
	}
	return netip.Prefix{}, fmt.Errorf("no ipv4 prefix found in AllowedIPs")
}

// CollectUsedIPsFromPeers 从 peers 列表中收集已使用的 IP
func CollectUsedIPsFromPeers(peers []*model.WGPeer, prefix netip.Prefix) map[netip.Addr]struct{} {
	used := make(map[netip.Addr]struct{})
	dbCount := 0

	for _, p := range peers {
		if p == nil {
			continue
		}
		cidr := strings.TrimSpace(p.ClientIP)
		if cidr == "" {
			continue
		}
		pr, err := netip.ParsePrefix(cidr)
		if err != nil {
			klog.V(2).InfoS("invalid client IP in DB", "peer_id", p.ID, "client_ip", cidr, "error", err)
			continue
		}
		ip := pr.Addr()
		if !ip.Is4() {
			continue
		}
		// 验证 IP 是否在分配范围内
		if !prefix.Contains(ip) {
			klog.V(2).InfoS("client IP outside allocation prefix", "peer_id", p.ID, "client_ip", ip, "prefix", prefix)
			continue
		}
		// 检查重复
		if _, exists := used[ip]; exists {
			klog.Warningf("duplicate IP found in DB: %s (peer_id: %s)", ip, p.ID)
		}
		used[ip] = struct{}{}
		dbCount++
	}

	klog.V(1).InfoS("collected used IPs", "total", len(used), "from_db", dbCount, "prefix", prefix)
	return used
}
