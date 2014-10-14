package focker

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

func (f *Focker) RunStager(writer io.Writer) error {
	prepareStagingFilesystem(f.directories)
	prepareStagingApp(f.directories.App(), f.directories.Staging())
	containerConfig := config.NewStageContainerConfig(f.directories)
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, containerConfig)
	DeleteContainer(writer, containerConfig.ContainerName)
	return stager.ValidateStagedApp(f.directories)
}

func (f *Focker) StageApp(writer io.Writer, buildpackDirOptional ...string) error {
	buildpackDir := f.directories.ContainerBuildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpackRunner := stager.NewBuildpackRunner(abs(buildpackDir))
	err := stager.RunBuildpack(writer, buildpackRunner)
	return err
}

func (f *Focker) RunRuntime(writer io.Writer) {
	prepareRuntimeFilesystem(f.directories.Droplet())
	containerConfig := config.NewRuntimeContainerConfig(f.directories.Droplet())
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	if docker.GetContainerId(cli, Stdout, stdoutpipe, containerConfig.ContainerName) != "" {
		fmt.Println("Deleting running runtime container...")
		f.StopRuntime(writer)
	}
	cli, Stdout, stdoutpipe = docker.GetNewClient()
	docker.RunConfiguredContainer(cli, Stdout, stdoutpipe, writer, containerConfig)
	fmt.Fprintln(writer, "Connect to your running application at http://localhost:8080/")
}

func (f *Focker) StopRuntime(writer io.Writer) {
	StopContainer(writer, "cloudfocker-runtime")
	DeleteContainer(writer, "cloudfocker-runtime")
}

func (f *Focker) BuildRuntimeImage(writer io.Writer) {
	prepareRuntimeFilesystem(f.directories.Droplet())
	containerConfig := config.NewRuntimeContainerConfig(f.directories.Droplet())
	cli, Stdout, stdoutpipe := docker.GetNewClient()
	docker.BuildRuntimeImage(cli, Stdout, stdoutpipe, writer, containerConfig)
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

func prepareStagingApp(appDir string, stagingDir string) {
	copyDir(appDir, stagingDir)
}

func copyDir(src string, dest string) {
	src = src + "/*"
	command := "cp -ra " + src + " " + dest
	if err := exec.Command("bash", "-c", command).Run(); err != nil {
		log.Fatalf("error copying from %s to %s : %s", src, dest, err)
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
	purgeHostDirectories(directories)

	if err := createHostDirectories(directories); err != nil {
		return err
	}

	return nil
}

func purgeHostDirectories(directories *config.Directories) {
	for _, dir := range directories.HostDirectoriesToClean() {
		os.RemoveAll(dir)
	}

	cleanTmpDirExceptCache(directories.Tmp())
}

func cleanTmpDirExceptCache(tmpDirName string) error {
	tmpDir, err := os.Open(tmpDirName)

	tmpDirContents, err := tmpDir.Readdirnames(0)
	for _, file := range tmpDirContents {
		if file != "cache" {
			os.RemoveAll(tmpDirName + "/" + file)
		}
	}
	return err
}

func createHostDirectories(directories *config.Directories) error {
	for _, dir := range directories.HostDirectories() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}