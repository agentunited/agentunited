package platforms

import (
	"errors"
	"sync"
)

var (
	ErrAdapterNotFound = errors.New("adapter not found")
	ErrAdapterExists   = errors.New("adapter already registered")
)

// Registry stores platform adapters
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]PlatformAdapter
}

// NewRegistry creates a new adapter registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]PlatformAdapter),
	}
}

// Register adds an adapter to the registry
func (r *Registry) Register(adapter PlatformAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	platform := adapter.Platform()
	if _, exists := r.adapters[platform]; exists {
		return ErrAdapterExists
	}

	r.adapters[platform] = adapter
	return nil
}

// Get retrieves an adapter by platform name
func (r *Registry) Get(platform string) (PlatformAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[platform]
	if !ok {
		return nil, ErrAdapterNotFound
	}
	return adapter, nil
}

// List returns all registered platform names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	platforms := make([]string, 0, len(r.adapters))
	for p := range r.adapters {
		platforms = append(platforms, p)
	}
	return platforms
}

// Unregister removes an adapter from the registry
func (r *Registry) Unregister(platform string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.adapters, platform)
}

// Global default registry
var defaultRegistry = NewRegistry()

// RegisterAdapter registers an adapter with the default registry
func RegisterAdapter(adapter PlatformAdapter) error {
	return defaultRegistry.Register(adapter)
}

// GetAdapter retrieves an adapter from the default registry
func GetAdapter(platform string) (PlatformAdapter, error) {
	return defaultRegistry.Get(platform)
}

// ListAdapters lists all adapters in the default registry
func ListAdapters() []string {
	return defaultRegistry.List()
}
