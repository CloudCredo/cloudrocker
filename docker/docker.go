package docker

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/cloudcredo/cloudfocker/config"

	"github.com/dotcloud/docker/api/client"
)

type DockerClient interface {
	CmdVersion(...string) error
	CmdImport(...string) error
	CmdRun(...string) error
	CmdStop(...string) error
	CmdRm(...string) error
	CmdKill(...string) error
	CmdPs(...string) error
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
	CopyFromPipeToPipe(stdout, writer)
	return nil
}

func ImportRootfsImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, url string) error {
	fmt.Fprintln(writer, "Bootstrapping Docker setup - this will take a few minutes...")
	go func() {
		err := cli.CmdImport(url, "cloudfocker-base")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(stdout, writer)
	return nil
}

func RunConfiguredContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, runConfig *config.RunConfig) error {
	fmt.Fprintln(writer, "Starting the CloudFocker container...")
	go func() {
		err := cli.CmdRun(ParseRunCommand(runConfig)...)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(stdout, writer)
	fmt.Fprintln(writer, "Started the CloudFocker container.")
	return nil
}

func StopContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Stopping the CloudFocker container...")
	go func() {
		err := cli.CmdStop(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(stdout, writer)
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func KillContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Killing the CloudFocker container...")
	go func() {
		err := cli.CmdKill(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(stdout, writer)
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func DeleteContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, name string) error {
	fmt.Fprintln(writer, "Deleting the CloudFocker container...")
	go func() {
		err := cli.CmdRm(name)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	CopyFromPipeToPipe(stdout, writer)
	fmt.Fprintln(writer, "Deleted container.")
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
		nil, stdoutPipe, nil, "unix", "/var/run/docker.sock", nil)
	return
}

func CopyFromPipeToPipe(reader *io.PipeReader, writer io.Writer) {
	rawBytes, _ := ioutil.ReadAll(reader)
	s := string(rawBytes[:])
	fmt.Fprint(writer, s)
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
