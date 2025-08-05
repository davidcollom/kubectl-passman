package main

import (
	"fmt"
	"os"

	"github.com/chrisns/kubectl-passman/internal/cli"
)

// VERSION populated at build time.
var VERSION = "0.0.0"

func main() {
	app := cli.NewApp(VERSION)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Exit with success code
	os.Exit(0)
}
