// Package keyring provides a provider implementation that utilises the system's keychain/keyring
// for securely storing and retrieving Kubernetes and application secrets. This package implements
// the provider.Provider interface, allowing integration with kubectl-passman for secret management.
package keyring

import (
	"sync"

	"github.com/chrisns/kubectl-passman/pkg/provider"
	keyring "github.com/zalando/go-keyring"
)

const (
	serviceName = "kubectl-passman"
)

// Provider provides common metadata methods for all keychain providers.
type Provider struct {
	mu sync.RWMutex
}

// Ensure Provider implements the provider.Provider interface (except Get/Set).
var _ provider.Provider = &Provider{}

// Name returns the name of the provider.
func (b *Provider) Name() string {
	return "keychain"
}

// Description returns a description of the provider.
func (b *Provider) Description() string {
	return "Use your systems keychain/keyring for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider.
func (b *Provider) Aliases() []string {
	return []string{"keyring", "kr"}
}

// Get returns an error for the base implementation - should be overridden.
func (b *Provider) Get(itemName string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return keyring.Get(serviceName, itemName)
}

// Set returns an error for the base implementation - should be overridden.
func (b *Provider) Set(itemName, secret string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return keyring.Set(serviceName, itemName, secret)
}

// Delete returns an error if the base implementation - should be overridden.
func (b *Provider) Delete(itemName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return keyring.Delete(serviceName, itemName)
}
