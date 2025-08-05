package cli

import (
	"encoding/json"
	"fmt"

	"github.com/chrisns/kubectl-passman/internal/registry"
	"github.com/chrisns/kubectl-passman/pkg/passman"
	"github.com/chrisns/kubectl-passman/pkg/provider"
	"github.com/urfave/cli"

	// Import all providers through the providers package
	_ "github.com/chrisns/kubectl-passman/internal/providers"
)

// App represents the CLI application
type App struct {
	app *cli.App
}

// NewApp creates a new CLI application
func NewApp(version string) *App {
	app := &App{
		app: cli.NewApp(),
	}

	app.setupCLI(version)
	return app
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

// Run starts the CLI application
func (a *App) Run(args []string) error {
	return a.app.Run(args)
}

func (a *App) cliHandler(c *cli.Context) error {
	handler := c.Command.Name
	itemName := c.Args().Get(0)
	secret := c.Args().Get(1)

	if itemName == "" {
		return cli.NewExitError("Please provide [item-name]", 1)
	}

	provider, exists := registry.GetProvider(handler)
	if !exists {
		return cli.NewExitError(fmt.Sprintf("Provider %s not found", handler), 1)
	}

	if secret != "" {
		return a.write(provider, itemName, secret)
	}
	return a.read(provider, itemName)
}

func (a *App) write(prov provider.Provider, itemName, secret string) error {
	validSecret, err := passman.FormatValidator(secret)
	if err != nil {
		return err
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
		return err
	}

	out, err := passman.FormatResponse(res)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}
