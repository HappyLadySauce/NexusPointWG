package options

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// WireGuardOptions contains configuration for WireGuard config management.
// All paths may be absolute; if user-dir is relative, it will be resolved under root-dir.
type WireGuardOptions struct {
	// RootDir is the WireGuard configuration root directory (default: /etc/wireguard).
	RootDir string `json:"root-dir" mapstructure:"root-dir"`

	// Interface is the server interface name (and config file base name), e.g. wg0 -> <rootDir>/wg0.conf.
	Interface string `json:"interface" mapstructure:"interface"`

	// UserDir is the directory to store generated user configuration files.
	// If relative, it is resolved under RootDir.
	UserDir string `json:"user-dir" mapstructure:"user-dir"`

	// Endpoint is the public endpoint exposed to clients, e.g. 10.10.10.10:51820.
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`

	// DNS is an optional DNS server for client configs.
	DNS string `json:"dns" mapstructure:"dns"`

	// DefaultAllowedIPs is the default AllowedIPs used in client configs (comma-separated CIDRs).
	// Example: "100.100.100.0/24,192.168.1.0/24"
	DefaultAllowedIPs string `json:"default-allowed-ips" mapstructure:"default-allowed-ips"`

	// ApplyMethod determines how to apply server config changes.
	// Supported: "systemctl", "none".
	ApplyMethod string `json:"apply-method" mapstructure:"apply-method"`

	// ServerIP is the server public IP for client endpoint (optional, auto-detected if empty)
	ServerIP string `json:"server_ip" mapstructure:"server_ip"`
}

func NewWireGuardOptions() *WireGuardOptions {
	return &WireGuardOptions{
		RootDir:           "/etc/wireguard",
		Interface:         "wg0",
		UserDir:           "user",
		Endpoint:          "",
		DNS:               "",
		DefaultAllowedIPs: "",
		ApplyMethod:       "systemctl",
	}
}

func (o *WireGuardOptions) Validate() []error {
	var errs []error
	if strings.TrimSpace(o.RootDir) == "" {
		errs = append(errs, fmt.Errorf("wireguard.root-dir is required"))
	}
	if strings.TrimSpace(o.Interface) == "" {
		errs = append(errs, fmt.Errorf("wireguard.interface is required"))
	}
	if strings.TrimSpace(o.Endpoint) == "" {
		errs = append(errs, fmt.Errorf("wireguard.endpoint is required"))
	}
	switch strings.ToLower(strings.TrimSpace(o.ApplyMethod)) {
	case "", "systemctl":
		// default
	case "none":
		// ok
	default:
		errs = append(errs, fmt.Errorf("wireguard.apply-method must be one of [systemctl, none]"))
	}
	return errs
}

func (o *WireGuardOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.RootDir, "wireguard.root-dir", o.RootDir, "WireGuard configuration root directory (default: /etc/wireguard)")
	fs.StringVar(&o.Interface, "wireguard.interface", o.Interface, "WireGuard server interface name (e.g. wg0). Server config file will be <root-dir>/<interface>.conf")
	fs.StringVar(&o.UserDir, "wireguard.user-dir", o.UserDir, "Directory for generated user configs (absolute or relative to root-dir)")
	fs.StringVar(&o.Endpoint, "wireguard.endpoint", o.Endpoint, "Public endpoint advertised to clients, e.g. 127.0.0.1:51820")
	fs.StringVar(&o.DNS, "wireguard.dns", o.DNS, "Optional DNS server for client configs, e.g. 1.1.1.1")
	fs.StringVar(&o.DefaultAllowedIPs, "wireguard.default-allowed-ips", o.DefaultAllowedIPs, "Default AllowedIPs for client configs (comma-separated CIDRs), e.g. 0.0.0.0/0,::/0")
	fs.StringVar(&o.ApplyMethod, "wireguard.apply-method", o.ApplyMethod, "How to apply server config changes: systemctl|none")
}

func (o *WireGuardOptions) ServerConfigPath() string {
	return filepath.Join(o.RootDir, o.Interface+".conf")
}

func (o *WireGuardOptions) ResolvedUserDir() string {
	if filepath.IsAbs(o.UserDir) {
		return o.UserDir
	}
	return filepath.Join(o.RootDir, o.UserDir)
}
