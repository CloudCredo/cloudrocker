package docker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudcredo/cloudrocker/config"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/docker/docker/api/client"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/archiver/compressor"
)

type DockerClient interface {
	CmdImport(...string) error
	CmdRun(...string) error
	CmdStop(...string) error
	CmdRm(...string) error
	CmdKill(...string) error
	CmdPs(...string) error
	CmdBuild(...string) error
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

func GetNewClient() (
	cli *client.DockerCli, stdout *io.PipeReader, stdoutPipe *io.PipeWriter) {
	stdout, stdoutPipe = io.Pipe()
	cli = client.NewDockerCli(
		nil, stdoutPipe, nil, nil, "unix", "/var/run/docker.sock", nil)
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
