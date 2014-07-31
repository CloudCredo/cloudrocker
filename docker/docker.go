package docker

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/utils"

	"github.com/dotcloud/docker/api/client"
)

type DockerClient interface {
	CmdVersion(...string) error
	CmdImport(...string) error
	CmdBuild(...string) error
	CmdRun(...string) error
	CmdStop(...string) error
	CmdRm(...string) error
	CmdKill(...string) error
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

func RunContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer) error {
	fmt.Fprintln(writer, "Running the CloudFocker container...")
	go func() {
		err := cli.CmdRun("-d", "--publish=8080:8080", "--name=cloudfocker-container", "cloudfocker:latest")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	PrintToStdout(stdout, stdoutPipe, "Finished starting the CloudFocker container", writer)
	fmt.Fprintln(writer, "Connect to your running application at http://localhost:8080/")
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
	PrintToStdout(stdout, stdoutPipe, "Finished starting the CloudFocker container", writer)
	fmt.Fprintln(writer, "Started the CloudFocker container.")
	return nil
}

func StopContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer) error {
	fmt.Fprintln(writer, "Stopping the CloudFocker container...")
	go func() {
		err := cli.CmdStop("cloudfocker-container")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	PrintToStdout(stdout, stdoutPipe, "Finished stopping the CloudFocker container", writer)
	fmt.Fprintln(writer, "Stopped your application.")
	return nil
}

func KillContainer(cli DockerClient, stdout *io.PipeReader, stdoutPipe *io.PipeWriter, writer io.Writer) error {
	fmt.Fprintln(writer, "Killing the CloudFocker container...")
	go func() {
		err := cli.CmdKill("cloudfocker-container")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err = closeWrap(stdout, stdoutPipe); err != nil {
			log.Fatalf("Error: %s", err)
		}
	}()
	PrintToStdout(stdout, stdoutPipe, "Finished killing the CloudFocker container", writer)
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
	PrintToStdout(stdout, stdoutPipe, "Finished deleting the CloudFocker container", writer)
	fmt.Fprintln(writer, "Deleted container.")
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
