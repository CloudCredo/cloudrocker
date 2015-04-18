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
	CmdVersion(...string) error
	CmdImport(...string) error
	CmdRun(...string) error
	CmdStop(...string) error
	CmdRm(...string) error
	CmdKill(...string) error
	CmdPs(...string) error
	CmdBuild(...string) error
}

func PrintVersion(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer) error {
	fmt.Fprintln(writer, "Checking Docker version")
	go func() {
		err := cli.CmdVersion()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
	return nil
}

func ImportRootfsImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, url string) error {
	fmt.Fprintln(writer, "Bootstrapping Docker setup - this will take a few minutes...")
	go func() {
		err := cli.CmdImport(url, "cloudrocker-base")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(writer, stdout)
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

func BuildRuntimeImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, containerConfig *config.ContainerConfig) error {
	fmt.Fprintln(writer, "Creating image configuration...")
	compressor := compressor.NewTgz()
	compressor.Compress(containerConfig.DropletDir+"/app/", containerConfig.DropletDir+"/droplet.tgz")
	WriteRuntimeDockerfile(containerConfig)
	fmt.Fprintln(writer, "Creating image...")
	go func() {
		err := cli.CmdBuild(containerConfig.DropletDir)
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

//A few of functions stolen from Deis dockercliuitls! Thanks guys
func GetNewClient() (
	cli *client.DockerCli, stdout *io.PipeReader, stdoutPipe *io.PipeWriter) {
	stdout, stdoutPipe = io.Pipe()
	cli = client.NewDockerCli(
		nil, stdoutPipe, nil, nil, "unix", "/var/run/docker.sock", nil)
	return
}

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
