package focker

import (
	"io"

	"github.com/hatofmonkeys/cloudfocker/docker"
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
