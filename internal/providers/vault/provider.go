package vault

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/hashicorp/vault/api"
)

// Ensure Provider implements the provider.Provider interface
var _ provider.Provider = &Provider{}

// Provider implements the Provider interface for HashiCorp Vault
type Provider struct {
	client  *api.Client
	once    sync.Once
	initErr error
}

func init() {
	registry.Register(&Provider{})
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "vault"
}

// Description returns a description of the provider
func (p *Provider) Description() string {
	return "Use HashiCorp Vault for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider
func (p *Provider) Aliases() []string {
	return []string{"hcv"}
}

// Get retrieves a secret from Vault
func (p *Provider) Get(itemName string) (string, error) {
	if err := p.initClient(); err != nil {
		return "", err
	}

	// Parse the secret path and key
	path, key := p.parseSecretPath(itemName)

	// Read the secret from Vault
	secret, err := p.client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret from Vault: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("secret not found at path: %s", path)
	}

	// Extract the specific key from the secret data
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		// Try direct access for non-KV v2 secrets
		data = secret.Data
	}

	value, exists := data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in secret at path '%s'", key, path)
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("secret value is not a string")
	}

	return valueStr, nil
}

// Set stores a secret in Vault
func (p *Provider) Set(itemName, secret string) error {
	if err := p.initClient(); err != nil {
		return err
	}

	// Parse the secret path and key
	path, key := p.parseSecretPath(itemName)

	// Prepare the data for KV v2 engine (most common)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			key: secret,
		},
	}

	// Write the secret to Vault
	_, err := p.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret to Vault: %w", err)
	}

	return nil
}

// initClient initializes the Vault client lazily
func (p *Provider) initClient() error {
	p.once.Do(func() {
		// Check required environment variables
		vaultAddr := os.Getenv("VAULT_ADDR")
		vaultToken := os.Getenv("VAULT_TOKEN")

		if vaultAddr == "" || vaultToken == "" {
			p.initErr = fmt.Errorf("vault environment variables not set (VAULT_ADDR, VAULT_TOKEN)")
			return
		}

		// Create Vault config
		config := &api.Config{
			Address: vaultAddr,
		}

		// Create the client
		client, err := api.NewClient(config)
		if err != nil {
			p.initErr = fmt.Errorf("failed to create Vault client: %w", err)
			return
		}

		// Set the token
		client.SetToken(vaultToken)

		p.client = client
	})

	return p.initErr
}

// parseSecretPath parses an itemName into a Vault path and key
// Expected format: "secret/path/to/item:key" or just "secret/path/to/item" (defaults to "password" key)
func (p *Provider) parseSecretPath(itemName string) (string, string) {
	parts := strings.SplitN(itemName, ":", 2)
	path := parts[0]
	key := "password" // default key

	if len(parts) == 2 {
		key = parts[1]
	}

	return path, key
}
