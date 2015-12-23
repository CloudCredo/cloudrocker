package docker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudcredo/cloudrocker/config"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/archiver/compressor"
)

type GoDockerClient interface {
	Version() (*docker.Env, error)
	ImportImage(docker.ImportImageOptions) error
	BuildImage(docker.BuildImageOptions) error
}

type DockerClient interface {
	CmdImport(...string) error
	CmdRun(...string) error
	CmdStop(...string) error
	CmdRm(...string) error
	CmdKill(...string) error
	CmdPs(...string) error
	CmdBuild(...string) error
}

func PrintVersion(cli GoDockerClient, writer io.Writer) error {
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

func ImportRootfsImage(cli GoDockerClient, writer io.Writer, url string) error {
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

func RunConfiguredContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Starting the CloudRocker container...")
	if os.Getenv("DEBUG") == "true" {
		fmt.Println(ParseRunCommand(containerConfig))
	}
	go func() {
		err := cli.CmdRun(ParseRunCommand(containerConfig)...)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	fmt.Fprintln(writer, "Started the CloudRocker container.")
	return nil
}

func StopContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Stopping the CloudRocker container...")
	go func() {
		err := cli.CmdStop(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func KillContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Killing the CloudRocker container...")
	go func() {
		err := cli.CmdKill(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func DeleteContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Deleting the CloudRocker container...")
	go func() {
		err := cli.CmdRm(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	fmt.Fprintln(writer, "Deleted container.")
	return nil
}

func BuildBaseImage(cli GoDockerClient, writer io.Writer, containerConfig *config.ContainerConfig) error {
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

func BuildRuntimeImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Creating image configuration...")
	compressor := compressor.NewTgz()
	compressor.Compress(containerConfig.DropletDir+"/app/", containerConfig.DropletDir+"/droplet.tgz")
	WriteRuntimeDockerfile(containerConfig)
	fmt.Fprintln(writer, "Creating image...")
	go func() {
		err := cli.CmdBuild(`--tag="`+containerConfig.DstImageTag+`"`, containerConfig.DropletDir)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	fmt.Fprintln(writer, "Created image.")
	return nil
}

func GetContainerId(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, containerName string) (containerId string) {
	go func() {
		err := cli.CmdPs()
		if err != nil {
			log.Fatalf("getContainerId %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("getContainerId %s", err)
		}
	}()
	reader := bufio.NewReader(stdout)
	for {
		cmdBytes, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(cmdBytes, containerName) {
			containerId = strings.Fields(cmdBytes)[0]
			return
		}
	}
	return
}

func GetNewClient() (cli *docker.Client) {
	cli, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	return
}

//A few of functions stolen from Deis dockercliuitls! Thanks guys
func CopyFromPipeToPipe(outputPipe io.Writer, inputPipe *io.PipeReader) {
	scanner := bufio.NewScanner(inputPipe)
	for scanner.Scan() {
		fmt.Fprintln(outputPipe, scanner.Text())
	}
}

func closeWrap(args ...io.Closer) error {
	e := false
	ret := fmt.Errorf("Error closing elements")
	for _, c := range args {
		if err := c.Close(); err != nil {
			e = true
			ret = fmt.Errorf("%s\n%s", ret, err)
		}
	}
	if e {
		return ret
	}
	return nil
}
