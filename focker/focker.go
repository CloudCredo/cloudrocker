package focker

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hatofmonkeys/cloudfocker/buildpack"
	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/docker"
	df "github.com/hatofmonkeys/cloudfocker/dockerfile"
	"github.com/hatofmonkeys/cloudfocker/stager"
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

func (f Focker) RunContainer(writer io.Writer) {
	f.BuildImage(writer)
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.RunContainer(cli, Stdout, stdoutpipe, writer)
}

func (f Focker) StopContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.StopContainer(cli, Stdout, stdoutpipe, writer, name)
	f.DeleteContainer(writer, name)
}

func (Focker) DeleteContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.DeleteContainer(cli, Stdout, stdoutpipe, writer, name)
}

func (Focker) AddBuildpack(writer io.Writer, url string, buildpackDirOptional ...string) {
	buildpackDir := utils.Cloudfockerhome() + "/buildpacks"
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Add(writer, url, abs(buildpackDir))
}

func (f Focker) RunStager(writer io.Writer, appDir string) {
	prepareStagingFilesystem(utils.Cloudfockerhome())
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	runConfig := config.NewStageRunConfig(abs(appDir))
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, runConfig)
	f.DeleteContainer(writer, runConfig.ContainerName)
}

func (Focker) StageApp(writer io.Writer, buildpackDirOptional ...string) error {
	buildpackDir := "/tmp/cloudfockerbuildpacks"
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpackRunner := stager.NewBuildpackRunner(abs(buildpackDir))
	err := stager.RunBuildpack(writer, buildpackRunner)
	return err
}

func cloudFockerfileLocation() (location string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	location = pwd + "/CloudFockerfile"
	return
}

func prepareStagingFilesystem(cloudfockerhome string) {
	if err := utils.CreateAndCleanAppDirs(cloudfockerhome); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := utils.AtLeastOneBuildpackIn(cloudfockerhome + "/buildpacks"); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := utils.CopyFockerBinaryToOwnDir(cloudfockerhome); err != nil {
		log.Fatalf(" %s", err)
	}
}

func abs(relative string) string {
	absolute, err := filepath.Abs(relative)
	if err != nil {
		log.Fatalf(" %s", err)
	}
	return absolute
}
