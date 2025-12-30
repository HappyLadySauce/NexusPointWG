package store

var (
	client Factory
)

// Factory defines the iam platform storage interface.
type Factory interface {
	Users() UserStore
	WGPeers() WGPeerStore
	IPPools() IPPoolStore
	IPAllocations() IPAllocationStore
	Close() error
}

// Client return the store client instance.
func Client() Factory {
	return client
}

// SetClient set the iam store client.
func SetClient(factory Factory) {
	client = factory
}
