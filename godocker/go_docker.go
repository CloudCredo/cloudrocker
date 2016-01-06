package godocker

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/archiver/compressor"
	"github.com/cloudcredo/cloudrocker/config"
)

type DockerClient interface {
	Version() (*docker.Env, error)
	ImportImage(docker.ImportImageOptions) error
	BuildImage(docker.BuildImageOptions) error
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	RemoveContainer(docker.RemoveContainerOptions) error
	StopContainer(containerName string, timeout uint) error
	CreateContainer(docker.CreateContainerOptions) (*docker.Container, error)
	StartContainer(string, *docker.HostConfig) error
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
	options := docker.ImportImageOptions{
		Source:       url,
		Repository:   "cloudrocker-raw",
		OutputStream: writer,
	}
	err := cli.ImportImage(options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return nil
}

func BuildBaseImage(cli DockerClient, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Creating image configuration...")
	WriteBaseImageDockerfile(containerConfig)
	fmt.Fprintln(writer, "Creating image...")
	options := docker.BuildImageOptions{
		Name:         containerConfig.DstImageTag,
		ContextDir:   containerConfig.BaseConfigDir,
		Dockerfile:   "/Dockerfile",
		OutputStream: writer,
	}
	err := cli.BuildImage(options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Created base image.")
	return nil
}

func BuildRuntimeImage(cli DockerClient, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Creating image configuration...")
	compressor := compressor.NewTgz()
	compressor.Compress(containerConfig.DropletDir+"/app/", containerConfig.DropletDir+"/droplet.tgz")
	WriteRuntimeDockerfile(containerConfig)
	fmt.Fprintln(writer, "Creating image...")
	options := docker.BuildImageOptions{
		Name:         containerConfig.DstImageTag,
		ContextDir:   containerConfig.DropletDir,
		Dockerfile:   "/Dockerfile",
		OutputStream: writer,
	}
	err := cli.BuildImage(options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Created runtime image.")
	return nil
}

func GetContainerID(cli DockerClient, containerName string) (containerID string) {
	options := docker.ListContainersOptions{
		All: true,
	}
	containers, err := cli.ListContainers(options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	for _, container := range containers {
		if container.Names[0] == "/"+containerName {
			return container.ID
		}
	}
	return ""
}

func DeleteContainer(cli DockerClient, writer io.Writer, containerName string) error {
	fmt.Fprintln(writer, "Deleting the CloudRocker container...")
	containerID := GetContainerID(cli, containerName)
	if containerID == "" {
		log.Fatalf("Error: No such container: %s", containerName)
	}
	options := docker.RemoveContainerOptions{
		ID:    containerID,
		Force: true,
	}
	err := cli.RemoveContainer(options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Deleted container.")
	return nil
}

func StopContainer(cli DockerClient, writer io.Writer, containerName string) error {
	fmt.Fprintln(writer, "Stopping the CloudRocker container...")
	containerID := GetContainerID(cli, containerName)
	if containerID == "" {
		log.Fatalf("Error: No such container: %s", containerName)
	}
	err := cli.StopContainer(containerID, 10)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func RunConfiguredContainer(cli DockerClient, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Starting the CloudRocker container...")
	var createOptions = ParseCreateContainerOptions(containerConfig)
	if os.Getenv("DEBUG") == "true" {
		fmt.Println(createOptions.Name)
		fmt.Println(createOptions.Config)
		fmt.Println(createOptions.HostConfig)
	}
	container, err := cli.CreateContainer(createOptions)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	err = cli.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Fprintln(writer, "Started the CloudRocker container.")
	return nil
}
