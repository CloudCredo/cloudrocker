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

func DockerVersion(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.PrintVersion(cli, Stdout, stdoutpipe, writer)
}

func ImportRootfsImage(writer io.Writer) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.ImportRootfsImage(cli, Stdout, stdoutpipe, writer, utils.GetRootfsUrl())
}

func StopContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.StopContainer(cli, Stdout, stdoutpipe, writer, name)
}

func DeleteContainer(writer io.Writer, name string) {
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.DeleteContainer(cli, Stdout, stdoutpipe, writer, name)
}

func (f *Focker) AddBuildpack(writer io.Writer, url string, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Add(writer, url, abs(buildpackDir))
}

func (f *Focker) DeleteBuildpack(writer io.Writer, bpack string, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Delete(writer, bpack, abs(buildpackDir))
}

func (f *Focker) ListBuildpacks(writer io.Writer, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.List(writer, abs(buildpackDir))
}

func (f *Focker) RunStager(writer io.Writer, appDir string) error {
	prepareStagingFilesystem(f.directories)
	stagingAppDir := prepareStagingApp(appDir, f.directories.Staging())
	runConfig := config.NewStageRunConfig(stagingAppDir, f.directories)
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, runConfig)
	DeleteContainer(writer, runConfig.ContainerName)
	return stager.ValidateStagedApp(f.directories)
}

func StageApp(writer io.Writer, buildpackDirOptional ...string) error {
	buildpackDir := "/tmp/cloudfockerbuildpacks"
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpackRunner := stager.NewBuildpackRunner(abs(buildpackDir))
	err := stager.RunBuildpack(writer, buildpackRunner)
	return err
}

func (f *Focker) RunRuntime(writer io.Writer) {
	prepareRuntimeFilesystem(f.directories.Droplet())
	runConfig := config.NewRuntimeRunConfig(f.directories.Droplet())
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	if docker.GetContainerId(cli, Stdout, stdoutpipe, runConfig.ContainerName) != "" {
		fmt.Println("Deleting running runtime container...")
		f.StopRuntime(writer)
	}
	cli, Stdout, stdoutpipe = docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, runConfig)
	fmt.Fprintln(writer, "Connect to your running application at http://localhost:8080/")
}

func (f *Focker) StopRuntime(writer io.Writer) {
	StopContainer(writer, "cloudfocker-runtime")
	DeleteContainer(writer, "cloudfocker-runtime")
}

func cloudFockerfileLocation() (location string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	location = pwd + "/CloudFockerfile"
	return
}

func prepareStagingFilesystem(directories *config.Directories) {
	if err := CreateAndCleanAppDirs(directories); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := buildpack.AtLeastOneBuildpackIn(directories.Buildpacks()); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := utils.CopyFockerBinaryToDir(directories.Focker()); err != nil {
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

func prepareRuntimeFilesystem(dropletDir string) {
	if err := utils.AddSoldierRunScript(dropletDir + "/app"); err != nil {
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

func CreateAndCleanAppDirs(directories *config.Directories) error {
	dirs := map[string]bool{
		directories.Buildpacks(): false,
		directories.Droplet():    true,
		directories.Cache():      false,
		directories.Result():     true,
		directories.Staging():    true,
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
