// Package registry provides a thread-safe global registry for managing provider implementations.
// It allows providers to be registered, retrieved by name or alias, and enumerated. The package
// also supports generating CLI commands for all registered providers using the urfave/cli package.
package registry

import (
	"sync"

	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"
)

// Registry holds all registered providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]provider.Provider
}

// Global registry instance.
var registry = &Registry{
	providers: make(map[string]provider.Provider),
}

// Register adds a provider to the global registry.
func Register(prov provider.Provider) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// Register the provider by its primary name
	registry.providers[prov.Name()] = prov

	// Register the provider by its aliases
	for _, alias := range prov.Aliases() {
		registry.providers[alias] = prov
	}
}

// GetProvider retrieves a provider by name.
func GetProvider(name string) (provider.Provider, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	p, exists := registry.providers[name]

	return p, exists
}

// GetAllProviders returns all registered providers.
func GetAllProviders() map[string]provider.Provider {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]provider.Provider)
	for name, p := range registry.providers {
		result[name] = p
	}

	return result
}

// GenerateCommands creates CLI commands from registered providers.
func GenerateCommands(handler func(*cli.Context) error) []cli.Command {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// Use a map to deduplicate providers (since aliases point to same provider)
	seen := make(map[provider.Provider]bool)

	commands := make([]cli.Command, 0, len(registry.providers))

	for _, provider := range registry.providers {
		if seen[provider] {
			continue
		}

		seen[provider] = true

		command := cli.Command{
			Name:      provider.Name(),
			Usage:     provider.Description(),
			Aliases:   provider.Aliases(),
			ArgsUsage: "[item-name]",
			Action:    handler,
		}
		commands = append(commands, command)
	}

	return commands
}
