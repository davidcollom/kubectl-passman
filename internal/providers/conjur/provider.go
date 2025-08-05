// Package conjur provides an implementation of the provider.Provider interface
// for integrating with CyberArk Conjur as a secrets management backend.
//
// This package allows storing and retrieving Kubernetes and application secrets
// using the Conjur API. The provider is registered automatically on package
// initialization and initialises the Conjur client lazily upon first use.
//
// Required environment variables for configuration:
//   - CONJUR_APPLIANCE_URL: The URL of the Conjur appliance.
//   - CONJUR_AUTHN_LOGIN: The Conjur authentication login.
//   - CONJUR_AUTHN_API_KEY: The API key for authentication.
//
// The Provider struct implements the provider.Provider interface, supporting
// methods for getting and setting secrets in Conjur. Errors are returned if
// required environment variables are not set or if client initialization fails.
package conjur

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/cyberark/conjur-api-go/conjurapi"
)

// Ensure Provider implements the provider.Provider interface.
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for Conjur.
type Provider struct {
	client *conjurapi.Client
	once   sync.Once
	err    error
}

func init() {
	// Always register the provider, but initialise client lazily
	registry.Register(&Provider{})
}

// ErrConjurNotConfigured is returned when Conjur environment variables are not set.
var ErrConjurNotConfigured = errors.New(
	"conjur environment variables not set (CONJUR_APPLIANCE_URL, CONJUR_AUTHN_LOGIN, CONJUR_AUTHN_API_KEY)",
)

// Name returns the name of the provider.
func (c *Provider) Name() string {
	return "conjur"
}

// Description returns a description of the provider.
func (c *Provider) Description() string {
	return "Use CyberArk Conjur for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider.
func (c *Provider) Aliases() []string {
	return []string{"ca"}
}

// Get retrieves a credential from Conjur.
func (c *Provider) Get(itemName string) (string, error) {
	if err := c.initClient(); err != nil {
		return "", err
	}

	data, err := c.client.RetrieveSecret(itemName)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Set stores a credential in Conjur.
func (c *Provider) Set(itemName, secret string) error {
	err := c.initClient()
	if err != nil {
		return err
	}

	return c.client.AddSecret(itemName, secret)
}

// initClient initialises the Conjur client on first use.
func (c *Provider) initClient() error {
	c.once.Do(func() {
		// Check if required environment variables are set
		if os.Getenv("CONJUR_APPLIANCE_URL") == "" || os.Getenv("CONJUR_AUTHN_LOGIN") == "" {
			c.err = ErrConjurNotConfigured

			return
		}

		config, err := conjurapi.LoadConfig()
		if err != nil {
			c.err = err

			return
		}

		client, err := conjurapi.NewClientFromEnvironment(config)
		if err != nil {
			c.err = fmt.Errorf("failed to create Conjur client: %w", err)

			return
		}

		c.client = client
	})

	return c.err
}
