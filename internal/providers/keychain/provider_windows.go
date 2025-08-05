//go:build windows
// +build windows

package keychain

import (
	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/zalando/go-keyring"
)

// Provider implements keychain operations for Windows using go-keyring
type Provider struct {
	Base // Embed the base provider for metadata methods
}

func init() {
	registry.Register(&Provider{})
}

// Get retrieves a credential from the Windows credential store
func (p *Provider) Get(itemName string) (string, error) {
	return keyring.Get(itemName, itemName)
}

// Set stores a credential in the Windows credential store
func (p *Provider) Set(itemName, secret string) error {
	return keyring.Set(itemName, itemName, secret)
}
