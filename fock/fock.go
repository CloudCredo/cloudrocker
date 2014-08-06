package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/hatofmonkeys/cloudfocker/focker"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "Fock"
	app.Usage = "Cloud Focker - fock the Cloud, run apps locally!"
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
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
			Usage: "download the Cloud Foundry base image",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.ImportRootfsImage(os.Stdout)
			},
		},
		{
			Name:  "up",
			Usage: "start the container for the application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				pwd, err := os.Getwd()
				if err != nil {
					log.Fatalf(" %s", err)
				} else {
					if err := focker.RunStager(os.Stdout, pwd); err != nil {
						log.Fatalf(" %s", err)
					}
				}
				focker.RunRuntime(os.Stdout)
			},
		},
		{
			Name:  "off",
			Usage: "stop the container and remove it",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.StopRuntime(os.Stdout)
			},
		},
		{
			Name:  "add-buildpack",
			Usage: "add-buildpack [URL] - add a buildpack from a GitHub URL",
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
			Usage: "delete-buildpack [BUILDPACK] - delete a buildpack the local system",
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
			Name:  "buildpacks",
			Usage: "show the buildpacks installed on the local system",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.ListBuildpacks(os.Stdout)
			},
		},
		{
			Name:  "stage",
			Usage: "stage an application",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				if internal := c.Args().First(); internal == "internal" {
					//this is focker being called inside the staging container
					focker.StageApp(os.Stdout)
				} else {
					//this is focker being called by the user, outside of the staging container
					pwd, err := os.Getwd()
					if err != nil {
						log.Fatalf(" %s", err)
					} else {
						if err := focker.RunStager(os.Stdout, pwd); err != nil {
							log.Fatalf(" %s", err)
						}
					}
				}
			},
		},
		{
			Name:  "run",
			Usage: "run the staged container",
			Action: func(c *cli.Context) {
				focker := focker.NewFocker()
				focker.RunRuntime(os.Stdout)
			},
		},
	}

	app.Run(os.Args)
}
