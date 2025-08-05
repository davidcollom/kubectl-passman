package registry

import (
	"sync"

	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"
)

// Registry holds all registered providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]provider.Provider
}

// Global registry instance
var registry = &Registry{
	providers: make(map[string]provider.Provider),
}

// Register adds a provider to the global registry
func Register(p provider.Provider) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// Register the provider by its primary name
	registry.providers[p.Name()] = p

	// Register the provider by its aliases
	for _, alias := range p.Aliases() {
		registry.providers[alias] = p
	}
}

// GetProvider retrieves a provider by name
func GetProvider(name string) (provider.Provider, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	p, exists := registry.providers[name]
	return p, exists
}

// GetAllProviders returns all registered providers
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

// GenerateCommands creates CLI commands from registered providers
func GenerateCommands(handler func(*cli.Context) error) []cli.Command {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	// Use a map to deduplicate providers (since aliases point to same provider)
	seen := make(map[provider.Provider]bool)
	var commands []cli.Command

	for _, p := range registry.providers {
		if seen[p] {
			continue
		}
		seen[p] = true

		command := cli.Command{
			Name:      p.Name(),
			Usage:     p.Description(),
			Aliases:   p.Aliases(),
			ArgsUsage: "[item-name]",
			Action:    handler,
		}
		commands = append(commands, command)
	}

	return commands
}
