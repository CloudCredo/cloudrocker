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

	}

	app.Commands = []cli.Command{
		{
			Name:  "docker",
			Usage: "print the local Docker info",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.DockerVersion(os.Stdout)
			},
		},
		{
			Name:  "this",
			Usage: "Download the Cloud Foundry base image",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.ImportRootfsImage(os.Stdout)
			},
		},
		{
			Name:  "dockerfile",
			Usage: "Create a dockerfile for the application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.WriteDockerfile(os.Stdout)
			},
		},
	}

	app.Run(os.Args)
}
