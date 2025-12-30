package spec

import "fmt"

// Resource represents a protected resource category.
// Keep these stable because they are referenced by policy.
type Resource string

const (
	ResourceUser     Resource = "user"
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
)
