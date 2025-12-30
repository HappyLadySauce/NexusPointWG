package spec

import "fmt"

// Resource represents a protected resource category.
// Keep these stable because they are referenced by policy.
type Resource string

const (
	ResourceUser     Resource = "user"
	ResourceWGPeer   Resource = "wg_peer"
	ResourceWGConfig Resource = "wg_config"
	ResourceIPPool   Resource = "ip_pool"
)

// Scope represents ownership scope of a resource.
// self: requester owns the resource; any: requester does not own it (or wants global scope).
type Scope string

const (
	ScopeSelf Scope = "self"
	ScopeAny  Scope = "any"
)

// Obj builds a canonical Casbin object string: "<resource>:<scope>".
func Obj(resource Resource, scope Scope) string {
	return fmt.Sprintf("%s:%s", resource, scope)
}

// Action represents an operation on an object.
// Use action names that map 1:1 to API intent so policy remains readable.
type Action string

const (
	// ---- user ----
	// Create: create a new user (admin-only when authenticated, public registration when not authenticated)
	ActionUserCreate Action = "user:create"
	// UpdateBasic: username/nickname/avatar/email
	ActionUserUpdateBasic Action = "user:update_basic"
	// UpdateSensitive: password/status/role (admin-only)
	ActionUserUpdateSensitive Action = "user:update_sensitive"
	// SoftDelete: set status=deleted (self-only)
	ActionUserSoftDelete Action = "user:soft_delete"
	// HardDelete: permanent removal (admin-only)
	ActionUserHardDelete Action = "user:hard_delete"
	// ChangePassword: change own password (self-only, requires old password verification)
	ActionUserChangePassword Action = "user:change_password"
	// List: list users (admin-only via policy)
	ActionUserList Action = "user:list"

	// ---- WireGuard peer ----
	// Create: create a new WireGuard peer
	ActionWGPeerCreate Action = "wg_peer:create"
	// Update: update an existing WireGuard peer
	ActionWGPeerUpdate Action = "wg_peer:update"
	// Delete: delete a WireGuard peer
	ActionWGPeerDelete Action = "wg_peer:delete"
	// List: list WireGuard peers
	ActionWGPeerList Action = "wg_peer:list"

	// ---- WireGuard config ----
	// Download: download WireGuard client configuration
	ActionWGConfigDownload Action = "wg_config:download"
	// Rotate: rotate WireGuard peer keys
	ActionWGConfigRotate Action = "wg_config:rotate"
	// Revoke: revoke WireGuard peer configuration
	ActionWGConfigRevoke Action = "wg_config:revoke"
	// Update: update WireGuard peer configuration
	ActionWGConfigUpdate Action = "wg_config:update"

	// ---- IP pool (admin-only) ----
	// Create: create a new IP pool
	ActionIPPoolCreate Action = "ip_pool:create"
	// Update: update an existing IP pool
	ActionIPPoolUpdate Action = "ip_pool:update"
	// Delete: delete an IP pool
	ActionIPPoolDelete Action = "ip_pool:delete"
	// List: list IP pools
	ActionIPPoolList Action = "ip_pool:list"
)
