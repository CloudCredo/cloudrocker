package godocker

import (
	"fmt"
	"io"
	"log"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cloudcredo/cloudrocker/config"
)

type DockerClient interface {
	Version() (*docker.Env, error)
	ImportImage(docker.ImportImageOptions) error
	BuildImage(docker.BuildImageOptions) error
}

func GetNewClient() (cli *docker.Client) {
	cli, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return
}

func PrintVersion(cli DockerClient, writer io.Writer) error {
	fmt.Fprintln(writer, "Checking Docker version")
	versionList, err := cli.Version()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Client OS/Arch: "+versionList.Get("Os")+"/"+versionList.Get("Arch"))
	fmt.Fprintln(writer, "Server version: "+versionList.Get("Version"))
	fmt.Fprintln(writer, "Server API version: "+versionList.Get("ApiVersion"))
	fmt.Fprintln(writer, "Server Go version: "+versionList.Get("GoVersion"))
	fmt.Fprintln(writer, "Server Git commit: "+versionList.Get("GitCommit"))
	return nil
}

func ImportRootfsImage(cli DockerClient, writer io.Writer, url string) error {
	fmt.Fprintln(writer, "Bootstrapping Docker setup - this will take a few minutes...")
	opts := docker.ImportImageOptions{
		Source:       url,
		Repository:   "cloudrocker-raw",
		Tag:          "cloudrocker-base:latest",
		OutputStream: writer,
	}
	err := cli.ImportImage(opts)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return nil
}

func BuildBaseImage(cli DockerClient, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Creating image configuration...")
	WriteBaseImageDockerfile(containerConfig)
	fmt.Fprintln(writer, "Creating image...")
	opts := docker.BuildImageOptions{
		Name:         containerConfig.DstImageTag,
		ContextDir:   containerConfig.BaseConfigDir,
		Dockerfile:   "/Dockerfile",
		OutputStream: writer,
	}
	err := cli.BuildImage(opts)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Created image.")
	return nil
}
