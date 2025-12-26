package spec

import (
	"path/filepath"
	"sync"

	casbin "github.com/casbin/casbin/v3"
	"k8s.io/klog/v2"
)

var (
	once     sync.Once
	enforcer *casbin.Enforcer
	initErr  error
)

// getEnforcer initializes and returns a singleton Casbin enforcer.
//
// Note: paths are relative to process working directory. If you later want config-driven
// paths, move these into options/config and plumb through here.
func getEnforcer() (*casbin.Enforcer, error) {
	once.Do(func() {
		modelPath := filepath.FromSlash("model.conf")
		policyPath := filepath.FromSlash("policy.csv")

		e, err := casbin.NewEnforcer(modelPath, policyPath)
		if err != nil {
			initErr = err
			return
		}
		enforcer = e
		klog.V(1).Infof("authz: casbin enforcer initialized (model=%s policy=%s)", modelPath, policyPath)
	})
	return enforcer, initErr
}

// Enforce checks whether subject can perform action on object.
// sub is typically a role ("admin"/"user") in the initial rollout.
func Enforce(sub, obj string, act Action) (bool, error) {
	e, err := getEnforcer()
	if err != nil {
		return false, err
	}
	return e.Enforce(sub, obj, string(act))
}


