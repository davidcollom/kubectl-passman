// Package gopass provides a provider implementation for storing and retrieving
// secrets using the gopass password manager. This package integrates with the
// kubectl-passman tool, allowing Kubernetes and application secrets to be managed
// securely via the gopass API.
//
// The Provider struct implements the provider.Provider interface, enabling
// seamless interaction with gopass for secret management operations such as
// retrieving and storing credentials.
//
// Usage of this package requires gopass to be properly installed and initialised
// on the host system.
package gopass

import (
	"context"
	"fmt"
	"sync"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

// Ensure Provider implements the provider.Provider interface.
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for gopass using the direct API.
type Provider struct {
	client  gopass.Store
	once    sync.Once
	initErr error
}

func init() {
	registry.Register(&Provider{})
}

// Name returns the name of the provider.
func (g *Provider) Name() string {
	return "gopass"
}

// Description returns a description of the provider.
func (g *Provider) Description() string {
	return "Use gopass for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider.
func (g *Provider) Aliases() []string {
	return []string{}
}

// Get retrieves a credential from gopass.
func (g *Provider) Get(itemName string) (string, error) {
	if err := g.initClient(); err != nil {
		return "", err
	}

	ctx := context.Background()

	secret, err := g.client.Get(ctx, itemName, "latest")
	if err != nil {
		return "", fmt.Errorf("failed to get secret from gopass: %w", err)
	}

	return string(secret.Bytes()), nil
}

// Set stores a credential in gopass.
func (g *Provider) Set(itemName, secret string) error {
	if err := g.initClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Create a new secret with the provided data
	sec := secrets.New()
	sec.SetPassword(secret)

	err := g.client.Set(ctx, itemName, sec)
	if err != nil {
		return fmt.Errorf("failed to set secret in gopass: %w", err)
	}

	return nil
}

// initClient initialises the gopass API client lazily.
func (g *Provider) initClient() error {
	g.once.Do(func() {
		ctx := context.Background()

		client, err := api.New(ctx)
		if err != nil {
			g.initErr = fmt.Errorf("failed to initialise gopass API: %w", err)

			return
		}

		g.client = client
	})

	return g.initErr
}
