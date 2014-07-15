package docker

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/dotcloud/docker/api/client"
)

type DockerClient interface {
	CmdVersion(...string) error
	CmdImport(...string) error
}

func PrintVersion(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer) error {
	fmt.Fprintln(writer, "Checking Docker version")
	go func() {
		err := cli.CmdVersion()
		if err != nil {
			fmt.Errorf(" %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			fmt.Errorf("Error: %s", err)
		}
	}()
	PrintToStdout(stdout, stdoutPipe, "Finished getting Docker version", writer)
	return nil
}

func ImportRootfsImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, url string) error {
	fmt.Fprintln(writer, "Bootstrapping Docker setup - this will take a few minutes...")
	go func() {
		err := cli.CmdImport(url, "cloudfocker-base")
		if err != nil {
			fmt.Errorf(" %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			fmt.Errorf("Error: %s", err)
		}
	}()
	PrintToStdout(stdout, stdoutPipe, "Finished bootstrapping", writer)
	return nil
}

//A few of functions stolen from Deis dockercliuitls! Thanks guys
func GetNewClient() (
	cli *client.DockerCli, stdout *io.PipeReader, stdoutPipe *io.PipeWriter) {
	stdout, stdoutPipe = io.Pipe()
	cli = client.NewDockerCli(
		nil, stdoutPipe, nil, "unix", "/var/run/docker.sock", nil)
	return
}

func PrintToStdout(stdout *io.PipeReader, stdoutPipe *io.PipeWriter, stoptag string, writer io.Writer) {
	for {
		if cmdBytes, err := bufio.NewReader(stdout).ReadString('\n'); err == nil {
			fmt.Fprint(writer, cmdBytes)
			if strings.Contains(cmdBytes, stoptag) == true {
				if err := closeWrap(stdout, stdoutPipe); err != nil {
					fmt.Errorf("Closewraps %s", err)
				}
			}
		} else {
			break
		}
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
