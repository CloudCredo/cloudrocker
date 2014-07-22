package docker

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hatofmonkeys/cloudfocker/utils"

	"github.com/dotcloud/docker/api/client"
)

type DockerClient interface {
	CmdVersion(...string) error
	CmdImport(...string) error
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
	PrintToStdout(stdout, stdoutPipe, "Finished getting Docker version", writer)
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
	PrintToStdout(stdout, stdoutPipe, "Finished bootstrapping", writer)
	return nil
}

func BuildImage(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer, cloudfockerfileLocation string) error {
	dockerfileLocation := cloudfockerfiletoDockerfile(cloudfockerfileLocation)

	fmt.Fprintln(writer, "Building the CloudFocker image...")
	go func() {
		err := cli.CmdBuild("--tag=cloudfocker", dockerfileLocation)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	defer os.RemoveAll(dockerfileLocation)

	PrintToStdout(stdout, stdoutPipe, "Finished building the CloudFocker image", writer)
	return nil
}

func cloudfockerfiletoDockerfile(cloudfockerfileLocation string) (dockerfileLocation string) {
	//copy the cffile to a tmp location Dockerfile
	dockerfileLocation, err := ioutil.TempDir(os.TempDir(), "cfockerbuilder")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	err = utils.Cp(cloudfockerfileLocation, dockerfileLocation+"/Dockerfile")
	if err != nil {
		log.Fatalf("Error: %s", err)
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

func PrintToStdout(stdout *io.PipeReader, stdoutPipe *io.PipeWriter, stoptag string, writer io.Writer) {
	for {
		if cmdBytes, err := bufio.NewReader(stdout).ReadString('\n'); err == nil {
			fmt.Fprint(writer, cmdBytes)
			if strings.Contains(cmdBytes, stoptag) == true {
				if err := closeWrap(stdout, stdoutPipe); err != nil {
					log.Fatalf("Error: Closewraps %s", err)
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
