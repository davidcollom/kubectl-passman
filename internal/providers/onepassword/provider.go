// Package onepassword provides an implementation of the provider.Provider
// interface for 1Password Connect,
// enabling secure storage and retrieval of Kubernetes authentication secrets
// using the 1Password Connect API. The Provider is registered automatically
// and initialises its client lazily on first use, reading configuration from
// the environment variables OP_CONNECT_HOST, OP_CONNECT_TOKEN, and OP_VAULT.
//
// Methods:
//   - Name: Returns the provider's name ("1password").
//   - Description: Returns a description of the provider.
//   - Aliases: Returns alternative names for the provider.
//   - Get: Retrieves a credential from 1Password by item name, searching for
//     fields labelled "credential", "password", or "secret", or the first
//     concealed field if none are found.
//   - Set: Stores a credential in 1Password as a new item with a concealed field.
//
// Errors:
//   - ErrNotConfigured: Returned if required environment variables are not set.
//   - ErrItemNotFound: Returned if the requested item is not found in the vault.
//   - ErrNoCredentialFieldFound: Returned if no suitable credential field is found in the item.
package onepassword

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
)

// Ensure Provider implements the provider.Provider interface.
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for 1Password Connect.
type Provider struct {
	client connect.Client
	vault  string
	once   sync.Once
	err    error
}

func init() {
	// Always register the provider, but initialise client lazily
	registry.Register(&Provider{})
}

var (
	ErrNotConfigured = errors.New(
		"1Password Connect environment variables not set (OP_CONNECT_HOST, OP_CONNECT_TOKEN, OP_VAULT)",
	)
	ErrItemNotFound           = errors.New("item not found in 1Password vault")
	ErrNoCredentialFieldFound = errors.New("no credential field found in 1Password item")
)

// Name returns the name of the provider.
func (p *Provider) Name() string {
	return "1password"
}

// Description returns a description of the provider.
func (p *Provider) Description() string {
	return "Use 1Password Connect for storing your kubernetes authentication secrets"
}

// Aliases returns alternative names for the provider.
func (p *Provider) Aliases() []string {
	return []string{"1pass", "op"}
}

// Get retrieves a credential from 1Password.
func (p *Provider) Get(itemName string) (string, error) {
	if err := p.initClient(); err != nil {
		return "", err
	}

	targetItem, err := p.findItemByName(itemName)
	if err != nil {
		return "", err
	}

	fullItem, err := p.client.GetItem(targetItem.ID, p.vault)
	if err != nil {
		return "", fmt.Errorf("failed to get item: %w", err)
	}

	return p.extractCredentialFromItem(fullItem)
}

// Set stores a credential in 1Password.
func (p *Provider) Set(itemName, secret string) error {
	if err := p.initClient(); err != nil {
		return err
	}

	// Create a new item with the credential
	item := &onepassword.Item{
		Title:    itemName,
		Category: "API_CREDENTIAL", // Use string constant instead
		Vault: onepassword.ItemVault{
			ID: p.vault,
		},
		Fields: []*onepassword.ItemField{
			{
				Label: "credential",
				Type:  "CONCEALED",
				Value: secret,
			},
		},
	}

	_, err := p.client.CreateItem(item, p.vault)

	return fmt.Errorf("failed to create item in 1Password: %w", err)
}

// initClient initialises the 1Password Connect client on first use.
func (p *Provider) initClient() error {
	p.once.Do(func() {
		connectHost := os.Getenv("OP_CONNECT_HOST")
		connectToken := os.Getenv("OP_CONNECT_TOKEN")
		vault := os.Getenv("OP_VAULT")

		if connectHost == "" || connectToken == "" || vault == "" {
			p.err = ErrNotConfigured

			return
		}

		p.client = connect.NewClient(connectHost, connectToken)
		p.vault = vault
	})

	return p.err
}

// findItemByName finds an item in the vault by its title.
func (p *Provider) findItemByName(itemName string) (*onepassword.Item, error) {
	items, err := p.client.GetItems(p.vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get items from 1Password: %w", err)
	}

	for i := range items {
		if items[i].Title == itemName {
			return &items[i], nil
		}
	}

	return nil, ErrItemNotFound
}

// extractCredentialFromItem extracts the credential value from an item's fields.
func (p *Provider) extractCredentialFromItem(item *onepassword.Item) (string, error) {
	// Look for a field labelled "credential" or similar
	if value := p.findCredentialField(item.Fields); value != "" {
		return value, nil
	}

	// If no specific credential field found, return the first concealed field
	if value := p.findConcealedField(item.Fields); value != "" {
		return value, nil
	}

	return "", ErrNoCredentialFieldFound
}

// findCredentialField looks for fields with credential-related labels.
func (p *Provider) findCredentialField(fields []*onepassword.ItemField) string {
	for _, field := range fields {
		if field.Label == "credential" || field.Label == "password" || field.Label == "secret" {
			return field.Value
		}
	}

	return ""
}

// findConcealedField finds the first concealed field in the item.
func (p *Provider) findConcealedField(fields []*onepassword.ItemField) string {
	for _, field := range fields {
		if field.Type == "CONCEALED" {
			return field.Value
		}
	}

	return ""
}
