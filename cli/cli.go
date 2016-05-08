package cli

import (
	cmd "github.com/codegangsta/cli"
	"github.com/lcaballero/again/start"
)

func NewCli() *cmd.App {
	app := cmd.NewApp()
	app.Name = "again"
	app.Version = "0.0.1"
	app.Usage = "Recursively watches .go files in a directory and restarts an executable."
	app.Commands = []cmd.Command{
		generateLookupCommand(),
	}
	return app
}

func generateLookupCommand() cmd.Command {
	return cmd.Command{
		Name:   "watch",
		Usage:  "Uses the current directory as root.",
		Action: start.Run,
	}
}

