package main

import (
	"github.com/lcaballero/again/cli"
	"os"
)

func main() {
	cli.NewCli().Run(os.Args)
}
