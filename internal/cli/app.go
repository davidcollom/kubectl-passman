// Package cli provides the command-line interface for the kubectl-passman application.
// It defines the App struct, which encapsulates the CLI application logic, including
// command registration, argument parsing, and interaction with credential providers.
//
// The CLI supports storing and retrieving kubeconfig credentials using various
// keychains or password managers, with providers auto-registered via the registry.
//
// Usage:
//
//	kubectl-passman [command] [item-name] [new-value(optional)]
//
// If new-value is provided, it writes to the item; otherwise, it reads the item.
//
// Main types and functions:
//   - App: Represents the CLI application.
//   - NewApp: Creates a new CLI application instance.
//   - Run: Starts the CLI application with provided arguments.
//
// Internal methods handle command setup, argument validation, provider lookup,
// and reading/writing secrets using the passman and provider packages.
package cli

import (
	"encoding/json"
	"fmt"

	// Import all providers through the providers package.
	_ "github.com/chrisns/kubectl-passman/internal/providers"
	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/passman"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"
)

// App represents the CLI application.
type App struct {
	app *cli.App
}

// NewApp creates a new CLI application.
func NewApp(version string) *App {
	app := &App{
		app: cli.NewApp(),
	}

	app.setupCLI(version)

	return app
}

// Run starts the CLI application.
func (a *App) Run(args []string) error {
	return a.app.Run(args) //nolint:wrapcheck
}

func (a *App) setupCLI(version string) {
	a.app.Name = "kubectl-passman"
	a.app.Usage = "Store kubeconfig credentials in keychains or password managers"
	a.app.Version = version
	a.app.Authors = []cli.Author{
		{
			Name:  "Chris Nesbitt-Smith",
			Email: "chris@cns.me.uk",
		},
	}
	a.app.Copyright = "(c) 2019 Chris Nesbitt-Smith"
	a.app.UsageText = `kubectl-passman [command] [item-name] [new-value(optional)]
	If new-value is provided it will write to the item, otherwise it will read`

	// Auto-generate commands from registered providers
	a.app.Commands = registry.GenerateCommands(a.cliHandler)
}

func (a *App) cliHandler(c *cli.Context) error {
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
