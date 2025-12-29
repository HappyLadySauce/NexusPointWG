package ip

import (
	"fmt"
	"net/netip"
	"strings"
)

// LastIPv4 计算子网的最后一个 IP（广播地址）
func LastIPv4(prefix netip.Prefix) netip.Addr {
	p := prefix.Masked()
	if !p.Addr().Is4() {
		return netip.Addr{}
	}
	base := p.Addr().As4()
	ones := p.Bits()
	hostBits := 32 - ones
	var n uint32
	n |= uint32(base[0]) << 24
	n |= uint32(base[1]) << 16
	n |= uint32(base[2]) << 8
	n |= uint32(base[3])
	if hostBits >= 32 {
		n |= ^uint32(0)
	} else if hostBits > 0 {
		n |= (uint32(1) << hostBits) - 1
	}
	b0 := byte(n >> 24)
	b1 := byte(n >> 16)
	b2 := byte(n >> 8)
	b3 := byte(n)
	return netip.AddrFrom4([4]byte{b0, b1, b2, b3})
}

// ParseFirstV4Prefix 解析第一个 IPv4 前缀
// Address may be comma-separated.
func ParseFirstV4Prefix(addressLine string) (netip.Prefix, netip.Addr, error) {
	parts := strings.Split(addressLine, ",")
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
			return prefix.Masked(), prefix.Addr(), nil
		}
	}
	return netip.Prefix{}, netip.Addr{}, fmt.Errorf("no ipv4 prefix found")
}
