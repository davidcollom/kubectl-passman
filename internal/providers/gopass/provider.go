package gopass

import (
	"io"
	"os/exec"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/provider"
)

// Ensure Provider implements the provider.Provider interface
var _ provider.Provider = &Provider{}

// Provider implements the provider.Provider interface for gopass
type Provider struct{}

func init() {
	registry.Register(&Provider{})
}

// Name returns the name of the provider
func (g *Provider) Name() string {
	return "gopass"
}

// Description returns a description of the provider
func (g *Provider) Description() string {
	return "Use gopass for storing your kubernetes and application secrets"
}

// Aliases returns alternative names for the provider
func (g *Provider) Aliases() []string {
	return []string{}
}

// Get retrieves a credential from gopass
func (g *Provider) Get(itemName string) (string, error) {
	out, err := exec.Command("gopass", "show", "--password", itemName).Output()
	return string(out), err
}

// Set stores a credential in gopass
func (g *Provider) Set(itemName, secret string) error {
	var stdin io.WriteCloser
	var err error

	cmd := exec.Command("gopass", "insert", "--force", itemName)

	stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	_, err = stdin.Write([]byte(secret))
	if err != nil {
		return err
	}

	err = stdin.Close()
	if err != nil {
		return err
	}

	_, err = cmd.Output()
	return err
}
