package spec

import (
	"bufio"
	_ "embed"
	"strings"
	"sync"

	casbin "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"k8s.io/klog/v2"
)

//go:embed model.conf
var modelConf []byte

//go:embed policy.csv
var policyCsv []byte

// stringAdapter is a simple adapter that loads policies from a string
type stringAdapter struct {
	policyText string
}

func (a *stringAdapter) LoadPolicy(m model.Model) error {
	scanner := bufio.NewScanner(strings.NewReader(a.policyText))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		persist.LoadPolicyLine(line, m)
	}
	return scanner.Err()
}

func (a *stringAdapter) SavePolicy(m model.Model) error {
	// Not implemented - policies are embedded, not writable
	return nil
}

func (a *stringAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	// Not implemented
	return nil
}

func (a *stringAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	// Not implemented
	return nil
}

func (a *stringAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	// Not implemented
	return nil
}

var (
	once     sync.Once
	enforcer *casbin.Enforcer
	initErr  error
)

// getEnforcer initializes and returns a singleton Casbin enforcer.
// Uses embedded files to avoid file path issues in Docker containers.
func getEnforcer() (*casbin.Enforcer, error) {
	once.Do(func() {
		// Load model from embedded string
		m, err := model.NewModelFromString(string(modelConf))
		if err != nil {
			klog.V(1).InfoS("failed to load casbin model from embedded file", "error", err)
			initErr = err
			return
		}

		// Create adapter from embedded policy string
		adapter := &stringAdapter{policyText: string(policyCsv)}

		// Create enforcer with model and adapter
		e, err := casbin.NewEnforcer(m, adapter)
		if err != nil {
			klog.V(1).InfoS("failed to initialize casbin enforcer", "error", err)
			initErr = err
			return
		}
		enforcer = e
		klog.V(1).InfoS("casbin enforcer initialized from embedded files")
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
