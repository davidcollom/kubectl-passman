package cli

import (
	"encoding/json"
	"fmt"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/passman"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"
)

func (a *App) providerHandler(c *cli.Context) error {
	handler := c.Command.Name
	itemName := c.Args().Get(0)
	secret := c.Args().Get(1)

	if itemName == "" {
		return cli.NewExitError("Please provide [item-name]", 1)
	}

	prov, exists := registry.GetProvider(handler)
	if !exists {
		return cli.NewExitError(fmt.Sprintf("Provider %s not found", handler), 1)
	}

	if secret != "" {
		return a.write(prov, itemName, secret)
	}

	return a.read(prov, itemName)
}

func (a *App) write(prov provider.Provider, itemName, secret string) error {
	validSecret, err := passman.FormatValidator(secret)
	if err != nil {
		return fmt.Errorf("failed to validate secret: %w", err)
	}

	return prov.Set(itemName, validSecret)
}

func (a *App) read(prov provider.Provider, itemName string) error {
	secret, err := prov.Get(itemName)
	if err != nil {
		return err
	}

	res := &passman.Response{}

	err = json.Unmarshal([]byte(secret), &res.Status)
	if err != nil {
		return fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	out, err := passman.FormatResponse(res)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	_, _ = a.app.Writer.Write([]byte(out + "\n")) //nolint:errcheck

	return nil
}
