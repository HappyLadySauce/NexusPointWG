package service

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/HappyLadySauce/NexusPointWG/internal/local"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
)

// extractServerInfo extracts server information from the server configuration.
func (w *wgSrv) extractServerInfo(ctx context.Context) (publicKey, serverIP, mtu string, allocationPrefix netip.Prefix, err error) {
	serverConfigStore := local.NewLocalStore().ServerConfigStore()
	configBytes, err := serverConfigStore.Read(ctx)
	if err != nil {
		return "", "", "", netip.Prefix{}, err
	}

	serverConfig, err := wgfile.ParseServerConfig(configBytes)
	if err != nil {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGServerConfigNotFound, "failed to parse server config: %v", err)
	}

	if serverConfig.Interface == nil {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGServerConfigNotFound, "server config has no Interface block")
	}

	// Extract private key and derive public key
	privateKey := strings.TrimSpace(serverConfig.Interface.PrivateKey)
	if privateKey == "" {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGServerPrivateKeyMissing, "server config has no PrivateKey")
	}
	publicKey, err = wgfile.DerivePublicKey(ctx, privateKey)
	if err != nil {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "failed to derive public key: %v", err)
	}

	// Extract MTU
	mtu = strings.TrimSpace(serverConfig.Interface.MTU)

	// Extract Address and parse it to get server IP and allocation prefix
	address := strings.TrimSpace(serverConfig.Interface.Address)
	if address == "" {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGServerAddressInvalid, "server config has no Address")
	}

	prefix, err := netip.ParsePrefix(address)
	if err != nil {
		return "", "", "", netip.Prefix{}, errors.WithCode(code.ErrWGServerAddressInvalid, "invalid Address format: %v", err)
	}

	// Server IP is the first IP in the prefix
	serverIP = prefix.Addr().String()

	return publicKey, serverIP, mtu, prefix, nil
}

// writeUserFiles writes user configuration files using structured rendering.
func (w *wgSrv) writeUserFiles(ctx context.Context, username string, peer *model.WGPeer, clientPriv, serverPub, mtu, endpoint string) error {
	cfg := config.Get()
	userFilesStore := local.NewLocalStore().UserConfigStore()

	// Derive client public key if not already set
	clientPub := peer.ClientPublicKey
	if clientPub == "" {
		var err error
		clientPub, err = wgfile.DerivePublicKey(ctx, clientPriv)
		if err != nil {
			return err
		}
	}

	// Determine AllowedIPs
	allowedIPs := strings.TrimSpace(peer.AllowedIPs)
	if allowedIPs == "" {
		if cfg != nil && cfg.WireGuard != nil {
			allowedIPs = strings.TrimSpace(cfg.WireGuard.DefaultAllowedIPs)
		}
		if allowedIPs == "" {
			allowedIPs = "0.0.0.0/0,::/0"
		}
	}

	// Determine DNS
	dns := strings.TrimSpace(peer.DNS)
	if dns == "" && cfg != nil && cfg.WireGuard != nil {
		dns = strings.TrimSpace(cfg.WireGuard.DNS)
	}

	// Build ClientConfig
	clientConfig := &wgfile.ClientConfig{
		Interface: &wgfile.InterfaceBlock{
			PrivateKey: clientPriv,
			Address:    peer.ClientIP,
			MTU:        mtu,
			DNS:        dns,
			Extra:      make(map[string]string),
		},
		Peer: &wgfile.PeerBlock{
			PublicKey:           serverPub,
			AllowedIPs:          allowedIPs,
			Endpoint:            endpoint,
			PersistentKeepalive: peer.PersistentKeepalive,
			Extra:               make(map[string]string),
		},
	}

	// Build ClientFiles
	clientFiles := &wgfile.ClientFiles{
		Config:     clientConfig,
		PrivateKey: clientPriv,
		PublicKey:  clientPub,
		Meta: &wgfile.ClientMeta{
			PeerID:      peer.ID,
			User:        username,
			DeviceName:  peer.DeviceName,
			ClientIP:    peer.ClientIP,
			Endpoint:    endpoint,
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
	}

	// Render all files
	renderedFiles := wgfile.RenderClientFiles(clientFiles)

	// Write each file
	for filename, content := range renderedFiles {
		if err := userFilesStore.Write(ctx, username, peer.ID, filename, content); err != nil {
			return err
		}
	}

	return nil
}

// readUserFiles reads user configuration files and parses them.
func (w *wgSrv) readUserFiles(ctx context.Context, username, peerID string) (*wgfile.ClientFiles, error) {
	userFilesStore := local.NewLocalStore().UserConfigStore()

	// Read all known files
	files := make(map[string][]byte)
	fileNames := []string{"peer.conf", "privatekey", "publickey", "meta.json"}

	for _, filename := range fileNames {
		content, err := userFilesStore.Read(ctx, username, peerID, filename)
		if err != nil {
			// Some files might not exist, continue
			continue
		}
		files[filename] = content
	}

	// Parse files
	return wgfile.ParseClientFiles(files)
}
