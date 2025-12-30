package local

// LocalStore defines storage operations for local file system.
type LocalStore interface {
	UserConfigStore() UserConfigStore
	ServerConfigStore() ServerConfigStore
}