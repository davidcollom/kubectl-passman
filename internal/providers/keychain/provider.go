package keychain

import (
	"fmt"

	"github.com/chrisns/kubectl-passman/pkg/provider"
)

const (
	AccessGroup = "github.com/chrisns/kubectl-passman"
	Service     = "kubectl-passman"
)

var ErrNotImplemented = fmt.Errorf("keyring provider not supported on this platform, please use the keychain provider for macOS or the gopass provider")

// Base provides common metadata methods for all keychain providers
type Base struct{}

// Ensure Base implements the provider.Provider interface (except Get/Set)
var _ provider.Provider = &Base{}

// Name returns the name of the provider
func (b *Base) Name() string {
	return "keychain"
}

// Description returns a description of the provider
func (b *Base) Description() string {
	return "Use your systems keychain/keyring for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider
func (b *Base) Aliases() []string {
	return []string{"keyring"}
}

// Get returns an error for the base implementation - should be overridden
func (b *Base) Get(itemName string) (string, error) {
	return "", ErrNotImplemented
}

// Set returns an error for the base implementation - should be overridden
func (b *Base) Set(itemName, secret string) error {
	return ErrNotImplemented
}
