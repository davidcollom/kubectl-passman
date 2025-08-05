package onepassword

import (
	"errors"
	"os"
	"sync"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
)

// Ensure Provider implements the provider.Provider interface
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for 1Password Connect
type Provider struct {
	client connect.Client
	vault  string
	once   sync.Once
	err    error
}

func init() {
	// Always register the provider, but initialize client lazily
	registry.Register(&Provider{})
}

// initClient initializes the 1Password Connect client on first use
func (p *Provider) initClient() error {
	p.once.Do(func() {
		connectHost := os.Getenv("OP_CONNECT_HOST")
		connectToken := os.Getenv("OP_CONNECT_TOKEN")
		vault := os.Getenv("OP_VAULT")

		if connectHost == "" || connectToken == "" || vault == "" {
			p.err = errors.New("1Password Connect environment variables not set (OP_CONNECT_HOST, OP_CONNECT_TOKEN, OP_VAULT)")
			return
		}

		p.client = connect.NewClient(connectHost, connectToken)
		p.vault = vault
	})
	return p.err
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "1password"
}

// Description returns a description of the provider
func (p *Provider) Description() string {
	return "Use 1Password Connect for storing your kubernetes authentication secrets"
}

// Aliases returns alternative names for the provider
func (p *Provider) Aliases() []string {
	return []string{"1pass", "op"}
}

// Get retrieves a credential from 1Password
func (p *Provider) Get(itemName string) (string, error) {
	if err := p.initClient(); err != nil {
		return "", err
	}

	// Get the item by title
	items, err := p.client.GetItems(p.vault)
	if err != nil {
		return "", err
	}

	var targetItem *onepassword.Item
	for _, item := range items {
		if item.Title == itemName {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		return "", errors.New("item not found in 1Password vault")
	}

	// Get the full item details
	fullItem, err := p.client.GetItem(targetItem.ID, p.vault)
	if err != nil {
		return "", err
	}

	// Look for a field labeled "credential" or similar
	for _, field := range fullItem.Fields {
		if field.Label == "credential" || field.Label == "password" || field.Label == "secret" {
			return field.Value, nil
		}
	}

	// If no specific credential field found, return the first concealed field
	for _, field := range fullItem.Fields {
		if field.Type == "CONCEALED" {
			return field.Value, nil
		}
	}

	return "", errors.New("no credential field found in 1Password item")
}

// Set stores a credential in 1Password
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
	return err
}
