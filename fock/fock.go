package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/cloudcredo/cloudfocker/focker"
)

func main() {
	app := cli.NewApp()
	app.Name = "fock"
	app.Version = "0.0.2"
	app.Usage = "Cloud Focker - fock the Cloud, run apps locally!"
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}

	app.Commands = []cli.Command{
		{
			Name:  "docker",
			Usage: "print the local Docker info",
			Action: func(c *cli.Context) {
				focker.DockerVersion(os.Stdout)
			},
		},
		{
			Name:  "this",
			Usage: "download the Cloud Foundry base image",
			Action: func(c *cli.Context) {
				focker.ImportRootfsImage(os.Stdout)
			},
		},
		{
			Name:  "up",
			Usage: "stage and run the application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if err := focker.RunStager(os.Stdout); err != nil {
					log.Fatalf(" %s", err)
				}
				focker.RunRuntime(os.Stdout)
			},
		},
		{
			Name:  "build",
			Usage: "build a runnable image of the application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if err := focker.RunStager(os.Stdout); err != nil {
					log.Fatalf(" %s", err)
				}
				focker.BuildRuntimeImage(os.Stdout)
			},
		},
		{
			Name:  "off",
			Usage: "stop the application container and remove it",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.StopRuntime(os.Stdout)
			},
		},
		{
			Name:  "buildpacks",
			Usage: "show the buildpacks installed on the local system",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.ListBuildpacks(os.Stdout)
			},
		},
		{
			Name:  "add-buildpack",
			Usage: "add-buildpack [URL] - add a buildpack from a GitHub URL to the local system",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if url := c.Args().First(); url != "" {
					focker.AddBuildpack(os.Stdout, url)
				} else {
					fmt.Println("Please supply a GitHub URL to download")
				}
			},
		},
		{
			Name:  "delete-buildpack",
			Usage: "delete-buildpack [BUILDPACK] - delete a buildpack from the local system",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if buildpack := c.Args().First(); buildpack != "" {
					focker.DeleteBuildpack(os.Stdout, buildpack)
				} else {
					fmt.Println("Please supply a buildpack to delete")
				}
			},
		},
		{
			Name:  "stage",
			Usage: "only execute the staging phase for the application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if internal := c.Args().First(); internal == "internal" {
					//this is focker being called inside the staging container
					focker.StageApp(os.Stdout)
				} else {
					//this is focker being called by the user, outside of the staging container
					if err := focker.RunStager(os.Stdout); err != nil {
						log.Fatalf(" %s", err)
					}
				}
			},
		},
		{
			Name:  "run",
			Usage: "only run the current staged application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.RunRuntime(os.Stdout)
			},
		},
	}

	app.Run(os.Args)
}
