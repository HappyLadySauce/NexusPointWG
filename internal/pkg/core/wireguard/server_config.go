package wireguard

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// ServerConfigManager manages WireGuard server configuration file.
type ServerConfigManager struct {
	configPath           string
	applyMethod          string
	mu                   sync.RWMutex
	serverPublicKeyCache string // Cached server public key
}

// NewServerConfigManager creates a new server configuration manager.
func NewServerConfigManager(configPath, applyMethod string) *ServerConfigManager {
	return &ServerConfigManager{
		configPath:  configPath,
		applyMethod: applyMethod,
	}
}

// ServerConfig represents the complete WireGuard server configuration.
type ServerConfig struct {
	Interface *InterfaceConfig
	Peers     []*ServerPeerConfig
}

// InterfaceConfig represents the [Interface] section of the server configuration.
type InterfaceConfig struct {
	PrivateKey string
	Address    string
	ListenPort int
	DNS        string
	MTU        int
	PreUp      string
	PostUp     string
	PreDown    string
	PostDown   string
	SaveConfig bool
}

// ReadServerConfig reads and parses the server configuration file.
func (m *ServerConfigManager) ReadServerConfig() (*ServerConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Open(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.WithCode(code.ErrWGServerConfigNotFound, "server config file not found: %s", m.configPath)
		}
		return nil, errors.Wrap(err, "failed to open server config file")
	}
	defer file.Close()

	config := &ServerConfig{
		Interface: &InterfaceConfig{},
		Peers:     make([]*ServerPeerConfig, 0),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentPeer *ServerPeerConfig
	var interfaceComments []string
	var peerComments []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle comments
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimPrefix(line, "#")
			comment = strings.TrimSpace(comment)
			if currentSection == "[Peer]" && currentPeer != nil {
				peerComments = append(peerComments, comment)
			} else if currentSection == "[Interface]" {
				interfaceComments = append(interfaceComments, comment)
			}
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.Trim(line, "[]")
			currentSection = "[" + sectionName + "]"

			if sectionName == "Peer" {
				// Save previous peer if exists
				if currentPeer != nil {
					if len(peerComments) > 0 {
						currentPeer.Comment = strings.Join(peerComments, "; ")
						peerComments = nil
					}
					config.Peers = append(config.Peers, currentPeer)
				}
				// Start new peer
				currentPeer = &ServerPeerConfig{}
			}
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "[Interface]":
			switch key {
			case "PrivateKey":
				config.Interface.PrivateKey = value
			case "Address":
				config.Interface.Address = value
			case "ListenPort":
				if port, err := strconv.Atoi(value); err == nil {
					config.Interface.ListenPort = port
				}
			case "DNS":
				config.Interface.DNS = value
			case "MTU":
				if mtu, err := strconv.Atoi(value); err == nil {
					config.Interface.MTU = mtu
				}
			case "PreUp":
				config.Interface.PreUp = value
			case "PostUp":
				config.Interface.PostUp = value
			case "PreDown":
				config.Interface.PreDown = value
			case "PostDown":
				config.Interface.PostDown = value
			case "SaveConfig":
				if save, err := strconv.ParseBool(value); err == nil {
					config.Interface.SaveConfig = save
				}
			}
		case "[Peer]":
			if currentPeer == nil {
				currentPeer = &ServerPeerConfig{}
			}
			switch key {
			case "PublicKey":
				currentPeer.PublicKey = value
			case "AllowedIPs":
				currentPeer.AllowedIPs = value
			case "PersistentKeepalive":
				if keepalive, err := strconv.Atoi(value); err == nil {
					currentPeer.PersistentKeepalive = keepalive
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read server config file")
	}

	// Add last peer if exists
	if currentPeer != nil {
		if len(peerComments) > 0 {
			currentPeer.Comment = strings.Join(peerComments, "; ")
		}
		config.Peers = append(config.Peers, currentPeer)
	}

	return config, nil
}

// GetServerPublicKey gets the server public key, generating it from the private key if needed.
// The result is cached to avoid frequent file reads.
func (m *ServerConfigManager) GetServerPublicKey() (string, error) {
	m.mu.RLock()
	if m.serverPublicKeyCache != "" {
		cached := m.serverPublicKeyCache
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	// Need to read config to get private key
	config, err := m.ReadServerConfig()
	if err != nil {
		return "", err
	}

	if config.Interface.PrivateKey == "" {
		return "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "server private key not found in config")
	}

	// Generate public key from private key
	publicKey, err := GeneratePublicKey(config.Interface.PrivateKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate server public key")
	}

	// Cache the result
	m.mu.Lock()
	m.serverPublicKeyCache = publicKey
	m.mu.Unlock()

	return publicKey, nil
}

// WriteServerConfig writes the server configuration to file.
func (m *ServerConfigManager) WriteServerConfig(config *ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create backup before writing
	backupPath := m.configPath + ".backup"
	if err := m.backupConfig(backupPath); err != nil {
		klog.V(1).InfoS("failed to create backup", "error", err)
		// Continue anyway, backup is not critical
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create config directory")
	}

	file, err := os.Create(m.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to create config file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write [Interface] section
	if config.Interface != nil {
		writer.WriteString("[Interface]\n")
		if config.Interface.PrivateKey != "" {
			writer.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.Interface.PrivateKey))
		}
		if config.Interface.Address != "" {
			writer.WriteString(fmt.Sprintf("Address = %s\n", config.Interface.Address))
		}
		if config.Interface.ListenPort > 0 {
			writer.WriteString(fmt.Sprintf("ListenPort = %d\n", config.Interface.ListenPort))
		}
		if config.Interface.DNS != "" {
			writer.WriteString(fmt.Sprintf("DNS = %s\n", config.Interface.DNS))
		}
		if config.Interface.MTU > 0 {
			writer.WriteString(fmt.Sprintf("MTU = %d\n", config.Interface.MTU))
		}
		if config.Interface.PreUp != "" {
			writer.WriteString(fmt.Sprintf("PreUp = %s\n", config.Interface.PreUp))
		}
		if config.Interface.PostUp != "" {
			writer.WriteString(fmt.Sprintf("PostUp = %s\n", config.Interface.PostUp))
		}
		if config.Interface.PreDown != "" {
			writer.WriteString(fmt.Sprintf("PreDown = %s\n", config.Interface.PreDown))
		}
		if config.Interface.PostDown != "" {
			writer.WriteString(fmt.Sprintf("PostDown = %s\n", config.Interface.PostDown))
		}
		if config.Interface.SaveConfig {
			writer.WriteString("SaveConfig = true\n")
		}
		writer.WriteString("\n")
	}

	// Write [Peer] sections
	for _, peer := range config.Peers {
		writer.WriteString(FormatServerPeerBlock(peer))
	}

	return nil
}

// backupConfig creates a backup of the current configuration file.
func (m *ServerConfigManager) backupConfig(backupPath string) error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	source, err := os.Open(m.configPath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = dest.ReadFrom(source)
	return err
}

// AddPeer adds a new peer to the server configuration.
func (m *ServerConfigManager) AddPeer(peer *ServerPeerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := m.readServerConfigUnsafe()
	if err != nil {
		return err
	}

	// Check if peer already exists
	for _, existingPeer := range config.Peers {
		if existingPeer.PublicKey == peer.PublicKey {
			return errors.WithCode(code.ErrIPAlreadyInUse, "peer with public key already exists")
		}
	}

	// Add new peer
	config.Peers = append(config.Peers, peer)

	return m.writeServerConfigUnsafe(config)
}

// RemovePeer removes a peer from the server configuration by public key.
func (m *ServerConfigManager) RemovePeer(publicKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := m.readServerConfigUnsafe()
	if err != nil {
		return err
	}

	// Find and remove peer
	newPeers := make([]*ServerPeerConfig, 0, len(config.Peers))
	found := false
	for _, peer := range config.Peers {
		if peer.PublicKey == publicKey {
			found = true
			continue
		}
		newPeers = append(newPeers, peer)
	}

	if !found {
		return errors.WithCode(code.ErrWGPeerNotFound, "peer with public key not found")
	}

	config.Peers = newPeers
	return m.writeServerConfigUnsafe(config)
}

// UpdatePeer updates an existing peer in the server configuration.
func (m *ServerConfigManager) UpdatePeer(publicKey string, peer *ServerPeerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := m.readServerConfigUnsafe()
	if err != nil {
		return err
	}

	// Find and update peer
	found := false
	for i, existingPeer := range config.Peers {
		if existingPeer.PublicKey == publicKey {
			// Update peer, but keep the public key from the parameter
			peer.PublicKey = publicKey
			config.Peers[i] = peer
			found = true
			break
		}
	}

	if !found {
		return errors.WithCode(code.ErrWGPeerNotFound, "peer with public key not found")
	}

	return m.writeServerConfigUnsafe(config)
}

// readServerConfigUnsafe reads the config without acquiring lock (caller must hold lock).
func (m *ServerConfigManager) readServerConfigUnsafe() (*ServerConfig, error) {
	file, err := os.Open(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.WithCode(code.ErrWGServerConfigNotFound, "server config file not found: %s", m.configPath)
		}
		return nil, errors.Wrap(err, "failed to open server config file")
	}
	defer file.Close()

	config := &ServerConfig{
		Interface: &InterfaceConfig{},
		Peers:     make([]*ServerPeerConfig, 0),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentPeer *ServerPeerConfig
	var peerComments []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			comment := strings.TrimPrefix(line, "#")
			comment = strings.TrimSpace(comment)
			if currentSection == "[Peer]" && currentPeer != nil {
				peerComments = append(peerComments, comment)
			}
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.Trim(line, "[]")
			currentSection = "[" + sectionName + "]"

			if sectionName == "Peer" {
				if currentPeer != nil {
					if len(peerComments) > 0 {
						currentPeer.Comment = strings.Join(peerComments, "; ")
						peerComments = nil
					}
					config.Peers = append(config.Peers, currentPeer)
				}
				currentPeer = &ServerPeerConfig{}
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "[Interface]":
			switch key {
			case "PrivateKey":
				config.Interface.PrivateKey = value
			case "Address":
				config.Interface.Address = value
			case "ListenPort":
				if port, err := strconv.Atoi(value); err == nil {
					config.Interface.ListenPort = port
				}
			case "DNS":
				config.Interface.DNS = value
			case "MTU":
				if mtu, err := strconv.Atoi(value); err == nil {
					config.Interface.MTU = mtu
				}
			case "PreUp":
				config.Interface.PreUp = value
			case "PostUp":
				config.Interface.PostUp = value
			case "PreDown":
				config.Interface.PreDown = value
			case "PostDown":
				config.Interface.PostDown = value
			case "SaveConfig":
				if save, err := strconv.ParseBool(value); err == nil {
					config.Interface.SaveConfig = save
				}
			}
		case "[Peer]":
			if currentPeer == nil {
				currentPeer = &ServerPeerConfig{}
			}
			switch key {
			case "PublicKey":
				currentPeer.PublicKey = value
			case "AllowedIPs":
				currentPeer.AllowedIPs = value
			case "PersistentKeepalive":
				if keepalive, err := strconv.Atoi(value); err == nil {
					currentPeer.PersistentKeepalive = keepalive
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read server config file")
	}

	if currentPeer != nil {
		if len(peerComments) > 0 {
			currentPeer.Comment = strings.Join(peerComments, "; ")
		}
		config.Peers = append(config.Peers, currentPeer)
	}

	return config, nil
}

// writeServerConfigUnsafe writes the config without acquiring lock (caller must hold lock).
func (m *ServerConfigManager) writeServerConfigUnsafe(config *ServerConfig) error {
	// Create backup
	backupPath := m.configPath + ".backup"
	if err := m.backupConfig(backupPath); err != nil {
		klog.V(1).InfoS("failed to create backup", "error", err)
	}

	// Create directory if needed
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create config directory")
	}

	file, err := os.Create(m.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to create config file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write [Interface] section
	if config.Interface != nil {
		writer.WriteString("[Interface]\n")
		if config.Interface.PrivateKey != "" {
			writer.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.Interface.PrivateKey))
		}
		if config.Interface.Address != "" {
			writer.WriteString(fmt.Sprintf("Address = %s\n", config.Interface.Address))
		}
		if config.Interface.ListenPort > 0 {
			writer.WriteString(fmt.Sprintf("ListenPort = %d\n", config.Interface.ListenPort))
		}
		if config.Interface.DNS != "" {
			writer.WriteString(fmt.Sprintf("DNS = %s\n", config.Interface.DNS))
		}
		if config.Interface.MTU > 0 {
			writer.WriteString(fmt.Sprintf("MTU = %d\n", config.Interface.MTU))
		}
		if config.Interface.PreUp != "" {
			writer.WriteString(fmt.Sprintf("PreUp = %s\n", config.Interface.PreUp))
		}
		if config.Interface.PostUp != "" {
			writer.WriteString(fmt.Sprintf("PostUp = %s\n", config.Interface.PostUp))
		}
		if config.Interface.PreDown != "" {
			writer.WriteString(fmt.Sprintf("PreDown = %s\n", config.Interface.PreDown))
		}
		if config.Interface.PostDown != "" {
			writer.WriteString(fmt.Sprintf("PostDown = %s\n", config.Interface.PostDown))
		}
		if config.Interface.SaveConfig {
			writer.WriteString("SaveConfig = true\n")
		}
		writer.WriteString("\n")
	}

	// Write [Peer] sections
	for _, peer := range config.Peers {
		writer.WriteString(FormatServerPeerBlock(peer))
	}

	// Clear server public key cache since config might have changed
	m.serverPublicKeyCache = ""

	return nil
}

// ApplyConfig applies the server configuration by reloading the WireGuard interface.
func (m *ServerConfigManager) ApplyConfig() error {
	if m.applyMethod == "none" {
		klog.V(2).InfoS("apply method is 'none', skipping config reload")
		return nil
	}

	// Extract interface name from config path (e.g., /etc/wireguard/wg0.conf -> wg0)
	baseName := filepath.Base(m.configPath)
	interfaceName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Execute systemctl reload
	cmd := exec.Command("systemctl", "reload", fmt.Sprintf("wg-quick@%s", interfaceName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.V(1).InfoS("failed to reload WireGuard config", "interface", interfaceName, "error", err, "output", string(output))
		return errors.WithCode(code.ErrWGApplyFailed, "failed to reload WireGuard config: %s", string(output))
	}

	klog.V(2).InfoS("WireGuard config reloaded successfully", "interface", interfaceName)
	return nil
}
