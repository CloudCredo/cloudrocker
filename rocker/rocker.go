package rocker

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudcredo/cloudrocker/buildpack"
	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/docker"
	"github.com/cloudcredo/cloudrocker/stager"
	"github.com/cloudcredo/cloudrocker/utils"
)

type Rocker struct {
	Stdout      *io.PipeReader
	directories *config.Directories
}

func NewRocker() *Rocker {
	return &Rocker{
		directories: config.NewDirectories(utils.CloudrockerHome()),
	}
}

func DockerVersion(writer io.Writer) {
	client := docker.GetNewClient()
	docker.PrintVersion(client, writer)
}

func (f *Rocker) ImportRootfsImage(writer io.Writer) {
	client := docker.GetNewClient()
	docker.ImportRootfsImage(client, writer, utils.GetRootfsUrl())
	f.BuildBaseImage(writer)
}

func (f *Rocker) BuildBaseImage(writer io.Writer) {
	createHostDirectories(f.directories)
	containerConfig := config.NewBaseContainerConfig(f.directories.BaseConfig())
	client := docker.GetNewClient()
	docker.BuildBaseImage(client, writer, containerConfig)
}

func StopContainer(writer io.Writer, name string) {
	client := docker.GetNewClient()
	docker.StopContainer(client, writer, name)
}

func DeleteContainer(writer io.Writer, name string) {
	client := docker.GetNewClient()
	docker.DeleteContainer(client, writer, name)
}

func (f *Rocker) AddBuildpack(writer io.Writer, url string, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Add(writer, url, abs(buildpackDir))
}

func (f *Rocker) DeleteBuildpack(writer io.Writer, bpack string, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.Delete(writer, bpack, abs(buildpackDir))
}

func (f *Rocker) ListBuildpacks(writer io.Writer, buildpackDirOptional ...string) {
	buildpackDir := f.directories.Buildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpack.List(writer, abs(buildpackDir))
}

func (f *Rocker) RunStager(writer io.Writer) error {
	prepareStagingFilesystem(f.directories)
	prepareStagingApp(f.directories.App(), f.directories.Staging())
	containerConfig := config.NewStageContainerConfig(f.directories)
	client := docker.GetNewClient()
	docker.RunStagingContainer(client, writer, containerConfig)
	DeleteContainer(writer, containerConfig.ContainerName)
	return stager.ValidateStagedApp(f.directories)
}

func (f *Rocker) StageApp(writer io.Writer, buildpackDirOptional ...string) error {
	buildpackDir := f.directories.ContainerBuildpacks()
	if len(buildpackDirOptional) > 0 {
		buildpackDir = buildpackDirOptional[0]
	}
	buildpackRunner := stager.NewBuildpackRunner(abs(buildpackDir))
	err := stager.RunBuildpack(writer, buildpackRunner)
	return err
}

func (f *Rocker) RunRuntime(writer io.Writer) {
	prepareRuntimeFilesystem(f.directories)
	containerConfig := config.NewRuntimeContainerConfig(f.directories.Droplet())
	client := docker.GetNewClient()
	if docker.GetContainerID(client, containerConfig.ContainerName) != "" {
		fmt.Println("Deleting running runtime container...")
		f.StopRuntime(writer)
	}
	client = docker.GetNewClient()
	docker.RunRuntimeContainer(client, writer, containerConfig)
	fmt.Fprintln(writer, "Connect to your running application at http://localhost:8080/")
}

func (f *Rocker) StopRuntime(writer io.Writer) {
	StopContainer(writer, "cloudrocker-runtime")
	DeleteContainer(writer, "cloudrocker-runtime")
}

func (f *Rocker) BuildRuntimeImage(writer io.Writer, destImageTagOptional ...string) {
	prepareRuntimeFilesystem(f.directories)
	containerConfig := config.NewRuntimeContainerConfig(f.directories.Droplet(), destImageTagOptional...)
	client := docker.GetNewClient()
	docker.BuildRuntimeImage(client, writer, containerConfig)
}

func cloudRockerfileLocation() (location string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	location = pwd + "/CloudRockerfile"
	return
}

func prepareStagingFilesystem(directories *config.Directories) {
	if err := CreateAndCleanAppDirs(directories); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := buildpack.AtLeastOneBuildpackIn(directories.Buildpacks()); err != nil {
		log.Fatalf(" %s", err)
	}
	if err := utils.CopyRockerBinaryToDir(directories.Rocker()); err != nil {
		log.Fatalf(" %s", err)
	}
}

func prepareStagingApp(appDir string, stagingDir string) {
	copyDir(appDir, stagingDir)
}

func copyDir(src string, dest string) {
	src = src + "/*"
	command := "shopt -s dotglob && cp -ra " + src + " " + dest
	if err := exec.Command("bash", "-c", command).Run(); err != nil {
		log.Fatalf("error copying from %s to %s : %s", src, dest, err)
	}
}

func prepareRuntimeFilesystem(directories *config.Directories) {
	tarPath, err := exec.LookPath("tar")
	if err != nil {
		log.Fatalf(" %s", err)
	}

	err = exec.Command(tarPath, "-xzf", directories.Tmp()+"/droplet", "-C", directories.Droplet()).Run()
	if err != nil {
		log.Fatalf(" %s", err)
	}

	if err := utils.AddLauncherRunScript(directories.Droplet() + "/app"); err != nil {
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
