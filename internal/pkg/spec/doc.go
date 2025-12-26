// Package spec defines specification vocabulary and helpers used across controllers/services.
//
// Design goals:
// - Make authorization intent explicit: (subject, object, action)
// - Keep controllers simple: build object scope (self/any) and call Enforce()
// - Avoid scattering role checks (e.g. role==admin) across handlers
//
// Object naming convention (recommended):
//
//	<resource>:<scope>
//
// Examples:
//
//	user:self, user:any
//	wg_peer:self, wg_peer:any
//	wg_config:self, wg_config:any
//
// Actions are intentionally stringly-typed to keep Casbin policy readable.
package spec
