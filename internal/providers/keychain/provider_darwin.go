//go:build darwin
// +build darwin

package keychain

import (
	"fmt"

	"github.com/chrisns/kubectl-passman/internal/registry"
)

// Provider implements keychain operations for macOS using go-keyring
type Provider struct {
	Base // Embed the base provider for metadata methods
}

func init() {
	registry.Register(&Provider{})
}

// Get retrieves a credential from the macOS keychain
func (p *Provider) Get(itemName string) (string, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(Service)
	query.SetAccessGroup(AccessGroup)
	query.SetAccount(clusterEndpoint)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", err
	}
	if len(results) != 1 {
		return "", fmt.Errorf("Multiple secrets for %s", clusterEndpoint)
	}
	return string(results[0].Data), nil

}

// Set stores a credential in the macOS keychain
func (p *Provider) Set(itemName, secret string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(Service)
	item.SetAccount(clusterEndpoint)
	item.SetLabel(clusterName)
	item.SetAccessGroup(AccessGroup)
	item.SetData([]byte(credentials))
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)

}
