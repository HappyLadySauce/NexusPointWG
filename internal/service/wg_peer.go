package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/ip"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

// WGPeerSrv defines the interface for WireGuard peer business logic.
type WGPeerSrv interface {
	CreatePeer(ctx context.Context, userID, deviceName, ipPoolID, clientIP, allowedIPs, dns, endpoint, clientPrivateKey string, persistentKeepalive *int) (*model.WGPeer, error)
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error)
	UpdatePeer(ctx context.Context, peer *model.WGPeer) error
	DeletePeer(ctx context.Context, id string) error
	ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error)
}

type wgPeerSrv struct {
	store store.Factory
}

// WGPeerSrv if implemented, then wgPeerSrv implements WGPeerSrv interface.
var _ WGPeerSrv = (*wgPeerSrv)(nil)

func newWGPeers(s *service) *wgPeerSrv {
	return &wgPeerSrv{store: s.store}
}

func (w *wgPeerSrv) CreatePeer(ctx context.Context, userID, deviceName, ipPoolID, clientIP, allowedIPs, dns, endpoint, clientPrivateKey string, persistentKeepalive *int) (*model.WGPeer, error) {
	// Get default IP pool if not specified
	if ipPoolID == "" {
		pools, _, err := w.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{
			Status: model.IPPoolStatusActive,
			Limit:  1,
		})
		if err != nil || len(pools) == 0 {
			return nil, errors.WithCode(code.ErrIPPoolNotFound, "no active IP pool found")
		}
		ipPoolID = pools[0].ID
	}

	// Allocate IP address
	allocator := ip.NewAllocator(w.store)
	var allocatedIP string
	if clientIP != "" {
		// Validate and use provided IP
		if err := allocator.ValidateAndAllocateIP(ctx, ipPoolID, clientIP); err != nil {
			return nil, err
		}
		allocatedIP = clientIP
	} else {
		// Auto-allocate IP
		var err error
		allocatedIP, err = allocator.AllocateIP(ctx, ipPoolID, "")
		if err != nil {
			return nil, err
		}
	}

	// Format IP as CIDR
	clientIPCIDR, err := ip.FormatIPAsCIDR(allocatedIP)
	if err != nil {
		return nil, err
	}

	// Generate key pair
	var privateKey, publicKey string
	if clientPrivateKey != "" {
		// Validate provided private key
		if err := wireguard.ValidatePrivateKey(clientPrivateKey); err != nil {
			return nil, err
		}
		privateKey = clientPrivateKey
		publicKey, err = wireguard.GeneratePublicKey(privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		// Auto-generate key pair
		privateKey, publicKey, err = wireguard.GenerateKeyPair()
		if err != nil {
			return nil, err
		}
	}

	// Generate peer ID
	peerID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate peer ID")
	}

	// Create peer model
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              userID,
		DeviceName:          deviceName,
		ClientPrivateKey:    privateKey,
		ClientPublicKey:     publicKey,
		ClientIP:            clientIPCIDR,
		AllowedIPs:          allowedIPs,
		DNS:                 dns,
		Endpoint:            endpoint,
		PersistentKeepalive: 25,
		Status:              model.WGPeerStatusActive,
		IPPoolID:            ipPoolID,
	}

	if persistentKeepalive != nil {
		peer.PersistentKeepalive = *persistentKeepalive
	}

	// Create IP allocation record
	allocationID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate allocation ID")
	}

	allocation := &model.IPAllocation{
		ID:        allocationID,
		IPPoolID:  ipPoolID,
		PeerID:    peerID,
		IPAddress: allocatedIP,
		Status:    model.IPAllocationStatusAllocated,
	}

	// Save to database
	if err := w.store.WGPeers().CreatePeer(ctx, peer); err != nil {
		return nil, err
	}

	if err := w.store.IPAllocations().CreateIPAllocation(ctx, allocation); err != nil {
		// Rollback: delete peer if allocation fails
		_ = w.store.WGPeers().DeletePeer(ctx, peerID)
		return nil, err
	}

	return peer, nil
}

func (w *wgPeerSrv) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.store.WGPeers().GetPeer(ctx, id)
}

func (w *wgPeerSrv) GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error) {
	return w.store.WGPeers().GetPeerByPublicKey(ctx, publicKey)
}

func (w *wgPeerSrv) UpdatePeer(ctx context.Context, peer *model.WGPeer) error {
	return w.store.WGPeers().UpdatePeer(ctx, peer)
}

func (w *wgPeerSrv) DeletePeer(ctx context.Context, id string) error {
	return w.store.WGPeers().DeletePeer(ctx, id)
}

func (w *wgPeerSrv) ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	return w.store.WGPeers().ListPeers(ctx, opt)
}
