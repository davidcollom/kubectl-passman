package main

import (
	"log"
	"os"

	"github.com/chrisns/kubectl-passman/internal/cli"
)

// VERSION populated at build time
var VERSION = "0.0.0"

func main() {
	app := cli.NewApp(VERSION)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
