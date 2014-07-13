package main

import (
	"github.com/codegangsta/cli"
	"github.com/hatofmonkeys/cloudfocker/focker"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "Focker"
	app.Usage = "Fock the Cloud, run apps locally!"
	app.Action = func(c *cli.Context) {
		focker := focker.NewFocker()
		focker.DockerVersion(os.Stdout)
	}

	app.Run(os.Args)
}
