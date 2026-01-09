package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

// BatchCreatePeers creates multiple WireGuard peers in a transaction.
// Note: This method only handles database operations. WireGuard config file updates
// and client config generation should be handled separately after batch creation succeeds.
func (w *wgPeerSrv) BatchCreatePeers(ctx context.Context, peers []*model.WGPeer) error {
	return w.store.WGPeers().BatchCreatePeers(ctx, peers)
}

// BatchUpdatePeers updates multiple WireGuard peers in a transaction.
// Note: This method only handles database operations. WireGuard config file updates
// and client config generation should be handled separately after batch update succeeds.
func (w *wgPeerSrv) BatchUpdatePeers(ctx context.Context, peers []*model.WGPeer) error {
	return w.store.WGPeers().BatchUpdatePeers(ctx, peers)
}

// BatchDeletePeers deletes multiple WireGuard peers by IDs in a transaction.
// Note: This method only handles database operations. WireGuard config file updates
// should be handled separately after batch deletion succeeds.
func (w *wgPeerSrv) BatchDeletePeers(ctx context.Context, ids []string) error {
	return w.store.WGPeers().BatchDeletePeers(ctx, ids)
}
