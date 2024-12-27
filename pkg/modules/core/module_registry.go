package core

import (
	"encoding/json"
	"fmt"
	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"sync"
)

type ModuleRegistry struct {
	modules map[string]base.Module
	deps    map[string][]string
	mu      sync.RWMutex
	Loader  base.ModuleLoader
}

func NewModuleRegistry(loader base.ModuleLoader) *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]base.Module),
		deps:    make(map[string][]string),
		Loader:  loader,
	}
}

type defaultLoader struct{}

func (l *defaultLoader) Load(path string) (base.Module, error) {
	// Implement module loading logic here
	return nil, fmt.Errorf("not implemented")
}

func (l *defaultLoader) LoadFromConfig(config base.ModuleConfig) (base.Module, error) {
	// Implement config-based loading logic here
	return nil, fmt.Errorf("not implemented")
}

func (r *ModuleRegistry) Register(module base.Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := module.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	if err := module.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize %s: %w", name, err)
	}

	r.modules[name] = module
	return nil
}

func (r *ModuleRegistry) RegisterWithDeps(module base.Module, deps []string) error {
	if err := r.Register(module); err != nil {
		return err
	}
	r.deps[module.Name()] = deps
	return r.resolveDeps(module.Name())
}

func (r *ModuleRegistry) resolveDeps(name string) error {
	deps := r.deps[name]
	for _, dep := range deps {
		if _, exists := r.modules[dep]; !exists {
			return fmt.Errorf("missing dependency %s for module %s", dep, name)
		}
	}
	return nil
}

func (r *ModuleRegistry) LoadFromConfig(config []byte) error {
	var configs []base.ModuleConfig
	if err := json.Unmarshal(config, &configs); err != nil {
		return err
	}

	for _, cfg := range configs {
		mod, err := r.Loader.LoadFromConfig(cfg)
		if err != nil {
			return err
		}
		if err := r.RegisterWithDeps(mod, cfg.DependsOn); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRegistry) Get(name string) (base.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mod, exists := r.modules[name]
	return mod, exists
}

func (r *ModuleRegistry) Terminate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mod, exists := r.modules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}

	if err := mod.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate %s: %w", name, err)
	}

	delete(r.modules, name)
	return nil
}

// pkg/modules/core/registry.go

func (r *ModuleRegistry) List() []ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]ModuleInfo, 0, len(r.modules))
	for name, mod := range r.modules {
		modules = append(modules, ModuleInfo{
			Name:    name,
			Status:  mod.GetState(),
			Deps:    r.deps[name],
			Version: mod.Version(),
		})
	}
	return modules
}

type ModuleInfo struct {
	Name    string           `json:"name"`
	Status  base.ModuleState `json:"status"`
	Deps    []string         `json:"dependencies,omitempty"`
	Version string           `json:"version"`
}
