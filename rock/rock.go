package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/cloudcredo/cloudrocker/rocker"
)

func main() {
	app := cli.NewApp()
	app.Name = "rock"
	app.Version = "0.0.2"
	app.Usage = "Cloud Rocker - rock the Cloud, run apps locally!"
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}

	app.Commands = []cli.Command{
		{
			Name:  "docker",
			Usage: "print the local Docker info",
			Action: func(c *cli.Context) {
				rocker.DockerVersion(os.Stdout)
			},
		},
		{
			Name:  "this",
			Usage: "download the Cloud Foundry base image",
			Action: func(c *cli.Context) {
				rocker.ImportRootfsImage(os.Stdout)
			},
		},
		{
			Name:  "up",
			Usage: "stage and run the application",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				if err := rocker.RunStager(os.Stdout); err != nil {
					log.Fatalf(" %s", err)
				}
				rocker.RunRuntime(os.Stdout)
			},
		},
		{
			Name:  "build",
			Usage: "build a runnable image of the application",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				if err := rocker.RunStager(os.Stdout); err != nil {
					log.Fatalf(" %s", err)
				}
				rocker.BuildRuntimeImage(os.Stdout)
			},
		},
		{
			Name:  "off",
			Usage: "stop the application container and remove it",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				rocker.StopRuntime(os.Stdout)
			},
		},
		{
			Name:  "buildpacks",
			Usage: "show the buildpacks installed on the local system",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				rocker.ListBuildpacks(os.Stdout)
			},
		},
		{
			Name:  "add-buildpack",
			Usage: "add-buildpack [URL] - add a buildpack from a GitHub URL to the local system",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				if url := c.Args().First(); url != "" {
					rocker.AddBuildpack(os.Stdout, url)
				} else {
					fmt.Println("Please supply a GitHub URL to download")
				}
			},
		},
		{
			Name:  "delete-buildpack",
			Usage: "delete-buildpack [BUILDPACK] - delete a buildpack from the local system",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				if buildpack := c.Args().First(); buildpack != "" {
					rocker.DeleteBuildpack(os.Stdout, buildpack)
				} else {
					fmt.Println("Please supply a buildpack to delete")
				}
			},
		},
		{
			Name:  "stage",
			Usage: "only execute the staging phase for the application",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				if internal := c.Args().First(); internal == "internal" {
					//this is rocker being called inside the staging container
					if err := rocker.StageApp(os.Stdout); err != nil {
						fmt.Printf(" %s", err)
					}
				} else {
					//this is rocker being called by the user, outside of the staging container
					if err := rocker.RunStager(os.Stdout); err != nil {
						log.Fatalf(" %s", err)
					}
				}
			},
		},
		{
			Name:  "run",
			Usage: "only run the current staged application",
			Action: func(c *cli.Context) {
				rocker := rocker.NewRocker()
				rocker.RunRuntime(os.Stdout)
			},
		},
	}

	app.Run(os.Args)
}
