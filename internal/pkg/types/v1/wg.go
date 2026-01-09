package v1

// CreateWGPeerRequest represents a request to create a WireGuard peer.
// swagger:model
type CreateWGPeerRequest struct {
	// Username is the username of the user who owns this peer (admin can specify, regular user uses their own username)
	// If provided, UserID will be ignored and looked up by username
	Username string `json:"username,omitempty" binding:"omitempty"`
	// UserID is the ID of the user who owns this peer (deprecated, use Username instead)
	UserID string `json:"user_id,omitempty" binding:"omitempty"`
	// DeviceName is the name of the device (e.g., "My Laptop", "iPhone")
	DeviceName string `json:"device_name" binding:"required,min=1,max=64"`
	// ClientIP is the IP address to assign to the client (optional, will be auto-allocated if not provided)
	// Format: IPv4 address without CIDR (e.g., "100.100.100.2")
	ClientIP string `json:"client_ip,omitempty" binding:"omitempty,ipv4"`
	// IPPoolID is the ID of the IP pool to allocate from (required if ClientIP is not provided)
	IPPoolID string `json:"ip_pool_id,omitempty" binding:"omitempty"`
	// AllowedIPs is the allowed IPs for the peer (comma-separated CIDRs, optional, uses server default if not provided)
	AllowedIPs string `json:"allowed_ips,omitempty" binding:"omitempty,cidr"`
	// DNS is the DNS server(s) for the client (comma-separated, optional, uses server default if not provided)
	DNS string `json:"dns,omitempty" binding:"omitempty"`
	// Endpoint is the server endpoint (optional, uses server default if not provided)
	Endpoint string `json:"endpoint,omitempty" binding:"omitempty,endpoint"`
	// PersistentKeepalive is the keepalive interval in seconds (optional, default 25)
	PersistentKeepalive *int `json:"persistent_keepalive,omitempty" binding:"omitempty,min=0,max=65535"`
	// ClientPrivateKey is the WireGuard private key (optional, will be auto-generated if not provided)
	ClientPrivateKey string `json:"client_private_key,omitempty" binding:"omitempty"`
}

// UpdateWGPeerRequest represents a request to update a WireGuard peer.
// swagger:model
type UpdateWGPeerRequest struct {
	// DeviceName is the name of the device
	DeviceName *string `json:"device_name,omitempty" binding:"omitempty,min=1,max=64"`
	// ClientIP is the IP address to assign to the client (IPv4 address without CIDR, e.g., "100.100.100.2")
	ClientIP *string `json:"client_ip,omitempty" binding:"omitempty,ipv4"`
	// IPPoolID is the ID of the IP pool to allocate from
	IPPoolID *string `json:"ip_pool_id,omitempty" binding:"omitempty"`
	// ClientPrivateKey is the WireGuard private key
	ClientPrivateKey *string `json:"client_private_key,omitempty" binding:"omitempty"`
	// AllowedIPs is the allowed IPs for the peer (comma-separated CIDRs)
	AllowedIPs *string `json:"allowed_ips,omitempty" binding:"omitempty,cidr"`
	// DNS is the DNS server(s) for the client (comma-separated)
	DNS *string `json:"dns,omitempty" binding:"omitempty"`
	// Endpoint is the server endpoint
	Endpoint *string `json:"endpoint,omitempty" binding:"omitempty,endpoint"`
	// PersistentKeepalive is the keepalive interval in seconds
	PersistentKeepalive *int `json:"persistent_keepalive,omitempty" binding:"omitempty,min=0,max=65535"`
	// Status is the peer status (active/disabled)
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
	// Username is the username of the user to bind this peer to (admin-only, sensitive operation)
	Username *string `json:"username,omitempty" binding:"omitempty"`
}

// WGPeerResponse represents a WireGuard peer response.
// swagger:model
type WGPeerResponse struct {
	ID                  string `json:"id"`
	UserID              string `json:"user_id"`
	Username            string `json:"username,omitempty"` // Populated when listing peers
	DeviceName          string `json:"device_name"`
	ClientPublicKey     string `json:"client_public_key"`
	ClientPrivateKey    string `json:"client_private_key,omitempty"` // Optional, sensitive information
	ClientIP            string `json:"client_ip"`
	AllowedIPs          string `json:"allowed_ips"`
	DNS                 string `json:"dns,omitempty"`
	Endpoint            string `json:"endpoint,omitempty"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	Status              string `json:"status"`
	IPPoolID            string `json:"ip_pool_id,omitempty"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

// WGPeerListResponse represents a paginated list of WireGuard peers.
// swagger:model
type WGPeerListResponse struct {
	Total int64            `json:"total"`
	Items []WGPeerResponse `json:"items"`
}

// CreateIPPoolRequest represents a request to create an IP pool.
// swagger:model
type CreateIPPoolRequest struct {
	// Name is the name of the IP pool
	Name string `json:"name" binding:"required,min=1,max=64"`
	// CIDR is the CIDR range for the IP pool (e.g., "100.100.100.0/24")
	CIDR string `json:"cidr" binding:"required,cidr"`
	// Routes is the routes (comma-separated CIDRs) for client AllowedIPs (optional)
	Routes string `json:"routes,omitempty" binding:"omitempty,cidr"`
	// DNS is the DNS servers (comma-separated) for client config (optional)
	DNS string `json:"dns,omitempty" binding:"omitempty,dnslist"`
	// Endpoint is the server endpoint (e.g., "10.10.10.10:51820") (optional)
	Endpoint string `json:"endpoint,omitempty" binding:"omitempty,endpoint"`
	// Description is a description of the IP pool
	Description string `json:"description,omitempty" binding:"omitempty,max=255"`
}

// UpdateIPPoolRequest represents a request to update an IP pool.
// swagger:model
type UpdateIPPoolRequest struct {
	// Name is the name of the IP pool
	Name *string `json:"name,omitempty" binding:"omitempty,min=1,max=64"`
	// CIDR is the CIDR range for the IP pool (e.g., "100.100.100.0/24")
	// Can only be modified when no IPs are allocated from this pool
	CIDR *string `json:"cidr,omitempty" binding:"omitempty,cidr"`
	// Routes is the routes (comma-separated CIDRs) for client AllowedIPs
	Routes *string `json:"routes,omitempty" binding:"omitempty,cidr"`
	// DNS is the DNS servers (comma-separated) for client config
	DNS *string `json:"dns,omitempty" binding:"omitempty,dnslist"`
	// Endpoint is the server endpoint (e.g., "10.10.10.10:51820")
	Endpoint *string `json:"endpoint,omitempty" binding:"omitempty,endpoint"`
	// Description is a description of the IP pool
	Description *string `json:"description,omitempty" binding:"omitempty,max=255"`
	// Status is the pool status (active/disabled)
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
}

// IPPoolResponse represents an IP pool response.
// swagger:model
type IPPoolResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CIDR        string `json:"cidr"`
	Routes      string `json:"routes,omitempty"`
	DNS         string `json:"dns,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// IPPoolListResponse represents a paginated list of IP pools.
// swagger:model
type IPPoolListResponse struct {
	Total int64            `json:"total"`
	Items []IPPoolResponse `json:"items"`
}

// AvailableIPsResponse represents a response containing available IP addresses.
// swagger:model
type AvailableIPsResponse struct {
	IPPoolID string   `json:"ip_pool_id"`
	CIDR     string   `json:"cidr"`
	IPs      []string `json:"ips"`
	Total    int      `json:"total"`
}

// GetServerConfigResponse represents a response containing server configuration.
// swagger:model
type GetServerConfigResponse struct {
	// Address is the server tunnel IP (e.g., "100.100.100.1/24")
	Address string `json:"address"`
	// ListenPort is the listening port (e.g., 51820)
	ListenPort int `json:"listen_port"`
	// PrivateKey is the server private key (sensitive information)
	PrivateKey string `json:"private_key"`
	// MTU is the Maximum Transmission Unit (e.g., 1420)
	MTU int `json:"mtu"`
	// PostUp is the PostUp command
	PostUp string `json:"post_up"`
	// PostDown is the PostDown command
	PostDown string `json:"post_down"`
	// PublicKey is the server public key (calculated from private key)
	PublicKey string `json:"public_key"`
	// ServerIP is the server public IP for client endpoint (optional, auto-detected if empty)
	ServerIP string `json:"server_ip"`
	// DNS is the DNS server for client configs (optional, comma-separated IP addresses)
	DNS string `json:"dns"`
}

// UpdateServerConfigRequest represents a request to update server configuration.
// swagger:model
type UpdateServerConfigRequest struct {
	// Address is the server tunnel IP (e.g., "100.100.100.1/24")
	Address *string `json:"address,omitempty" binding:"omitempty,cidr"`
	// ListenPort is the listening port
	ListenPort *int `json:"listen_port,omitempty" binding:"omitempty,min=1,max=65535"`
	// PrivateKey is the server private key
	PrivateKey *string `json:"private_key,omitempty" binding:"omitempty,wgprivatekey"`
	// MTU is the Maximum Transmission Unit
	MTU *int `json:"mtu,omitempty" binding:"omitempty,min=68,max=65535"`
	// PostUp is the PostUp command
	PostUp *string `json:"post_up,omitempty" binding:"omitempty,max=1000"`
	// PostDown is the PostDown command
	PostDown *string `json:"post_down,omitempty" binding:"omitempty,max=1000"`
	// ServerIP is the server public IP for client endpoint (optional, auto-detected if empty)
	ServerIP *string `json:"server_ip,omitempty" binding:"omitempty,ipv4"`
	// DNS is the DNS server for client configs (optional, comma-separated IP addresses)
	DNS *string `json:"dns,omitempty" binding:"omitempty,dnslist"`
}

// BatchCreateIPPoolsRequest represents a batch IP pool creation request.
// swagger:model
type BatchCreateIPPoolsRequest struct {
	// Items is the list of IP pools to create (max 50 items)
	Items []CreateIPPoolRequest `json:"items" binding:"required,min=1,max=50,dive"`
}

// BatchCreateIPPoolsResponse represents a batch IP pool creation response.
// swagger:model
type BatchCreateIPPoolsResponse struct {
	// Count is the number of IP pools created successfully
	Count int64 `json:"count"`
}

// BatchUpdateIPPoolItem represents a single IP pool update item in batch operation.
// swagger:model
type BatchUpdateIPPoolItem struct {
	// ID is the ID of the IP pool to update
	ID string `json:"id" binding:"required"`
	// UpdateIPPoolRequest contains the fields to update
	UpdateIPPoolRequest
}

// BatchUpdateIPPoolsRequest represents a batch IP pool update request.
// swagger:model
type BatchUpdateIPPoolsRequest struct {
	// Items is the list of IP pools to update (max 50 items)
	Items []BatchUpdateIPPoolItem `json:"items" binding:"required,min=1,max=50,dive"`
}

// BatchUpdateIPPoolsResponse represents a batch IP pool update response.
// swagger:model
type BatchUpdateIPPoolsResponse struct {
	// Count is the number of IP pools updated successfully
	Count int64 `json:"count"`
}

// BatchDeleteIPPoolsRequest represents a batch IP pool deletion request.
// swagger:model
type BatchDeleteIPPoolsRequest struct {
	// IDs is the list of IP pool IDs to delete (max 50 items)
	IDs []string `json:"ids" binding:"required,min=1,max=50,dive,required"`
}

// BatchDeleteIPPoolsResponse represents a batch IP pool deletion response.
// swagger:model
type BatchDeleteIPPoolsResponse struct {
	// Count is the number of IP pools deleted successfully
	Count int64 `json:"count"`
}

// BatchCreateWGPeersRequest represents a batch WireGuard peer creation request.
// swagger:model
type BatchCreateWGPeersRequest struct {
	// Items is the list of WireGuard peers to create (max 50 items)
	Items []CreateWGPeerRequest `json:"items" binding:"required,min=1,max=50,dive"`
}

// BatchCreateWGPeersResponse represents a batch WireGuard peer creation response.
// swagger:model
type BatchCreateWGPeersResponse struct {
	// Count is the number of WireGuard peers created successfully
	Count int64 `json:"count"`
}

// BatchUpdateWGPeerItem represents a single WireGuard peer update item in batch operation.
// swagger:model
type BatchUpdateWGPeerItem struct {
	// ID is the ID of the WireGuard peer to update
	ID string `json:"id" binding:"required"`
	// UpdateWGPeerRequest contains the fields to update
	UpdateWGPeerRequest
}

// BatchUpdateWGPeersRequest represents a batch WireGuard peer update request.
// swagger:model
type BatchUpdateWGPeersRequest struct {
	// Items is the list of WireGuard peers to update (max 50 items)
	Items []BatchUpdateWGPeerItem `json:"items" binding:"required,min=1,max=50,dive"`
}

// BatchUpdateWGPeersResponse represents a batch WireGuard peer update response.
// swagger:model
type BatchUpdateWGPeersResponse struct {
	// Count is the number of WireGuard peers updated successfully
	Count int64 `json:"count"`
}

// BatchDeleteWGPeersRequest represents a batch WireGuard peer deletion request.
// swagger:model
type BatchDeleteWGPeersRequest struct {
	// IDs is the list of WireGuard peer IDs to delete (max 50 items)
	IDs []string `json:"ids" binding:"required,min=1,max=50,dive,required"`
}

// BatchDeleteWGPeersResponse represents a batch WireGuard peer deletion response.
// swagger:model
type BatchDeleteWGPeersResponse struct {
	// Count is the number of WireGuard peers deleted successfully
	Count int64 `json:"count"`
}
