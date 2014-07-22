package focker

import (
	"io"
	"log"
	"os"

	"github.com/hatofmonkeys/cloudfocker/docker"
	df "github.com/hatofmonkeys/cloudfocker/dockerfile"
	"github.com/hatofmonkeys/cloudfocker/utils"
)

type Focker struct {
	Stdout *io.PipeReader
}

func NewFocker() *Focker {
	return &Focker{}
}

func (Focker) DockerVersion(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.PrintVersion(cli, Stdout, stdoutpipe, writer)
}

func (Focker) ImportRootfsImage(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.ImportRootfsImage(cli, Stdout, stdoutpipe, writer, utils.GetRootfsUrl())
}

func (Focker) WriteDockerfile(writer io.Writer) {
	dockerfile := df.NewDockerfile()
	dockerfile.Create()
	dockerfile.Write(writer)
}

func (Focker) BuildImage(writer io.Writer) {
	dockerfile := df.NewDockerfile()
	dockerfile.Create()
	dockerfile.Persist(cloudFockerfileLocation())
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.BuildImage(cli, Stdout, stdoutpipe, writer, cloudFockerfileLocation())
}

func cloudFockerfileLocation() (location string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	location = pwd + "/CloudFockerfile"
	return
}
