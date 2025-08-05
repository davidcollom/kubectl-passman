package keyring

import (
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/zalando/go-keyring"
)

const (
	Service = "kubectl-passman"
)

// Provider provides common metadata methods for all keychain providers
type Provider struct{}

// Ensure Provider implements the provider.Provider interface (except Get/Set)
var _ provider.Provider = &Provider{}

// Name returns the name of the provider
func (b *Provider) Name() string {
	return "keychain"
}

// Description returns a description of the provider
func (b *Provider) Description() string {
	return "Use your systems keychain/keyring for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider
func (b *Provider) Aliases() []string {
	return []string{"keyring", "kr"}
}

// Get returns an error for the base implementation - should be overridden
func (b *Provider) Get(itemName string) (string, error) {
	return keyring.Get(Service, itemName)
}

// Set returns an error for the base implementation - should be overridden
func (b *Provider) Set(itemName, secret string) error {
	return keyring.Set(Service, itemName, secret)
}
