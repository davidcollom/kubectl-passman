//go:build linux
// +build linux

package keychain

import (
	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/zalando/go-keyring"
)

// Provider implements keychain operations for Linux using go-keyring
type Provider struct {
	Base // Embed the base provider for metadata methods
}

func init() {
	registry.Register(&Provider{})
}

// Get retrieves a credential from the Linux keyring
func (p *Provider) Get(itemName string) (string, error) {
	return keyring.Get(itemName, itemName)
}

// Set stores a credential in the Linux keyring
func (p *Provider) Set(itemName, secret string) error {
	return keyring.Set(itemName, itemName, secret)
}
