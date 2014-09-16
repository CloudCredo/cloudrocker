package focker

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cloudcredo/cloudfocker/buildpack"
	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/docker"
	"github.com/cloudcredo/cloudfocker/stager"
	"github.com/cloudcredo/cloudfocker/utils"
)

type Focker struct {
	Stdout      *io.PipeReader
	directories *config.Directories
}

func NewFocker() *Focker {
	return &Focker{
		directories: config.NewDirectories(utils.CloudfockerHome()),
	}
}

func (Focker) DockerVersion(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.PrintVersion(cli, Stdout, stdoutpipe, writer)
}

func (Focker) ImportRootfsImage(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.ImportRootfsImage(cli, Stdout, stdoutpipe, writer, utils.GetRootfsUrl())
}

func (f Focker) StopContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.StopContainer(cli, Stdout, stdoutpipe, writer, name)
}

func (Focker) DeleteContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.DeleteContainer(cli, Stdout, stdoutpipe, writer, name)
}

func (focker *Focker) AddBuildpack(writer io.Writer, url string, buildpackDirOptional ...string) {
	buildpackDir := focker.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Add(writer, url, abs(buildpackDir))
}

func (focker *Focker) DeleteBuildpack(writer io.Writer, bpack string, buildpackDirOptional ...string) {
	buildpackDir := focker.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Delete(writer, bpack, abs(buildpackDir))
}

func (focker *Focker) ListBuildpacks(writer io.Writer, buildpackDirOptional ...string) {
	buildpackDir := focker.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.List(writer, abs(buildpackDir))
}

func (focker *Focker) RunStager(writer io.Writer, appDir string) error {
	prepareStagingFilesystem(utils.CloudfockerHome(), focker.directories)
	stagingAppDir := prepareStagingApp(appDir, utils.CloudfockerHome()+"/staging")
	runConfig := config.NewStageRunConfig(stagingAppDir, focker.directories)
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, runConfig)
	focker.DeleteContainer(writer, runConfig.ContainerName)
	return stager.ValidateStagedApp(utils.CloudfockerHome())
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

func (f Focker) RunRuntime(writer io.Writer) {
	prepareRuntimeFilesystem(utils.CloudfockerHome())
	runConfig := config.NewRuntimeRunConfig(utils.CloudfockerHome() + "/droplet")
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	if docker.GetContainerId(cli, Stdout, stdoutpipe, runConfig.ContainerName) != "" {
		fmt.Println("Deleting running runtime container...")
		f.StopRuntime(writer)
	}
	cli, Stdout, stdoutpipe = docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, runConfig)
	fmt.Fprintln(writer, "Connect to your running application at http://localhost:8080/")
}

func (f Focker) StopRuntime(writer io.Writer) {
	f.StopContainer(writer, "cloudfocker-runtime")
	f.DeleteContainer(writer, "cloudfocker-runtime")
}

func cloudFockerfileLocation() (location string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	location = pwd + "/CloudFockerfile"
	return
}

func prepareStagingFilesystem(cloudfockerHome string, directories *config.Directories) {
	if err := CreateAndCleanAppDirs(cloudfockerHome, directories); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := buildpack.AtLeastOneBuildpackIn(directories.Buildpacks()); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := utils.CopyFockerBinaryToOwnDir(cloudfockerHome); err != nil {
		log.Fatalf(" %s", err)
	}
}

func prepareStagingApp(appDir string, stagingDir string) string {
	copyDir(appDir, stagingDir)
	return abs(stagingDir) + "/" + path.Base(appDir)
}

func copyDir(src string, dest string) {
	if err := exec.Command("cp", "-a", src, dest).Run(); err != nil {
		log.Fatalf(" %s", err)
	}
}

func prepareRuntimeFilesystem(cloudfockerHome string) {
	if err := utils.AddSoldierRunScript(cloudfockerHome + "/droplet/app"); err != nil {
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

func CreateAndCleanAppDirs(cloudfockerHomeDir string, directories *config.Directories) error {
	dirs := map[string]bool{
		directories.Buildpacks():        false,
		cloudfockerHomeDir + "/droplet": true,
		cloudfockerHomeDir + "/cache":   false,
		cloudfockerHomeDir + "/result":  true,
		cloudfockerHomeDir + "/staging": true,
	}

	for dir, clean := range dirs {
		if clean {
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
		}
	}
	for dir, _ := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
