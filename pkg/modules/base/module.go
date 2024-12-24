package base

import (
	"encoding/json"
	"fmt"
	"hydap/pkg/crypto"
)

// pkg/modules/base/module.go

type Module interface {
	Name() string
	Initialize() error
	Terminate() error
	Signature() string
	HealthCheck() error
	State() ModuleState
	Version() string
}
type ModuleConfig struct {
	Name      string
	Version   string
	DependsOn []string
	Config    map[string]interface{}
}

// pkg/modules/base/module.go

type ModuleState int

const (
	StateUninitialized ModuleState = iota
	StateInitialized
	StateRunning
	StatePaused
	StateError
)

type ModuleMetadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
}

// BaseModule provides common functionality
type BaseModule struct {
	metadata ModuleMetadata
	state    ModuleState
	hasher   *crypto.Blake3Hasher
	config   json.RawMessage
}

func (b *BaseModule) State() ModuleState {
	return b.state
}

func (b *BaseModule) Version() string {
	return b.metadata.Version
}
func (b *BaseModule) Name() string {
	return b.metadata.Name
}

func (b *BaseModule) Initialize() error {
	b.state = StateInitialized
	b.hasher = crypto.NewBlake3()
	return nil
}

func (b *BaseModule) Terminate() error {
	b.state = StateUninitialized
	return nil
}

func (b *BaseModule) Signature() string {
	if b.hasher == nil {
		b.hasher = crypto.NewBlake3()
	}
	return b.hasher.HashToBase64([]byte(b.Name()))
}

func (b *BaseModule) HealthCheck() error {
	if b.state == StateError {
		return fmt.Errorf("module %s is in error state", b.Name())
	}
	if b.state == StateUninitialized {
		return fmt.Errorf("module %s is not initialized", b.Name())
	}
	return nil
}

// Registry interfaces
type ModuleRegistrar interface {
	Register(module Module) error
	Unregister(name string) error
	Get(name string) (Module, error)
	List() []string
}

type ModuleLoader interface {
	Load(path string) (Module, error)
	LoadFromConfig(config ModuleConfig) (Module, error)
}

func (s ModuleState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitialized:
		return "initialized"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

func CreateNewModule(metadata ModuleMetadata, config json.RawMessage) Module {
	return &BaseModule{
		metadata: metadata,
		config:   config,
	}
}

func NewModuleMetadata(name, version, description, author, license string) ModuleMetadata {
	return ModuleMetadata{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		License:     license,
	}
}
