package llm

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages LLM providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) error {
	if p == nil {
		return fmt.Errorf("cannot register nil provider")
	}
	name := p.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider already registered: %s", name)
	}

	r.providers[name] = p
	return nil
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}

// List returns all registered provider names (sorted).
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Has checks if a provider is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.providers[name]
	return ok
}

// Count returns the number of registered providers.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.providers[name]; !ok {
		return fmt.Errorf("provider not found: %s", name)
	}
	delete(r.providers, name)
	return nil
}

// DefaultRegistry is the global provider registry.
var DefaultRegistry = NewRegistry()

// Register adds a provider to the default registry.
func Register(p Provider) error {
	return DefaultRegistry.Register(p)
}

// Get returns a provider from the default registry.
func Get(name string) (Provider, error) {
	return DefaultRegistry.Get(name)
}

// List returns all provider names from the default registry.
func List() []string {
	return DefaultRegistry.List()
}
