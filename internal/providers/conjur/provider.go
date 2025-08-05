package conjur

import (
	"errors"
	"os"
	"sync"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
)

// Ensure Provider implements the provider.Provider interface
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for Conjur
type Provider struct {
	client *conjurapi.Client
	once   sync.Once
	err    error
}

func init() {
	// Always register the provider, but initialize client lazily
	registry.Register(&Provider{})
}

// initClient initializes the Conjur client on first use
func (c *Provider) initClient() error {
	c.once.Do(func() {
		// Check if required environment variables are set
		if os.Getenv("CONJUR_APPLIANCE_URL") == "" || os.Getenv("CONJUR_AUTHN_LOGIN") == "" {
			c.err = errors.New("conjur environment variables not set (CONJUR_APPLIANCE_URL, CONJUR_AUTHN_LOGIN, CONJUR_AUTHN_API_KEY)")
			return
		}

		config, err := conjurapi.LoadConfig()
		if err != nil {
			c.err = err
			return
		}

		client, err := conjurapi.NewClientFromKey(config,
			authn.LoginPair{
				Login:  os.Getenv("CONJUR_AUTHN_LOGIN"),
				APIKey: os.Getenv("CONJUR_AUTHN_API_KEY"),
			},
		)
		if err != nil {
			c.err = err
			return
		}

		c.client = client
	})
	return c.err
}

// Name returns the name of the provider
func (c *Provider) Name() string {
	return "conjur"
}

// Description returns a description of the provider
func (c *Provider) Description() string {
	return "Use CyberArk Conjur for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider
func (c *Provider) Aliases() []string {
	return []string{"ca"}
}

// Get retrieves a credential from Conjur
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

// Set stores a credential in Conjur
func (c *Provider) Set(itemName, secret string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.client.AddSecret(itemName, secret)
}
