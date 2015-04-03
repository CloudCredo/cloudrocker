package config

import (
	"github.com/cloudcredo/cloudrocker/utils"
)

type Directories struct {
	mounts map[string]Directory
	app    string
}

type Directory struct {
	HostDirectory      string
	ContainerDirectory string
}

func NewDirectories(cloudFockerHomeDir string) *Directories {
	directories := &Directories{
		mounts: map[string]Directory{
			"home":       Directory{cloudFockerHomeDir, ""},
			"buildpacks": Directory{cloudFockerHomeDir + "/buildpacks", "/cloudfockerbuildpacks"},
			"droplet":    Directory{cloudFockerHomeDir + "/tmp/droplet", ""},
			"result":     Directory{cloudFockerHomeDir + "/tmp/result", ""},
			"cache":      Directory{cloudFockerHomeDir + "/tmp/cache", ""},
			"focker":     Directory{cloudFockerHomeDir + "/focker", "/focker"},
			"staging":    Directory{cloudFockerHomeDir + "/staging", "/app"},
			"tmp":        Directory{cloudFockerHomeDir + "/tmp", "/tmp"},
		},
		app: utils.Pwd(),
	}
	return directories
}

func (directories *Directories) Home() string {
	return directories.mounts["home"].HostDirectory
}

func (directories *Directories) Buildpacks() string {
	return directories.mounts["buildpacks"].HostDirectory
}

func (directories *Directories) ContainerBuildpacks() string {
	return directories.mounts["buildpacks"].ContainerDirectory
}

func (directories *Directories) Droplet() string {
	return directories.mounts["droplet"].HostDirectory
}

func (directories *Directories) Result() string {
	return directories.mounts["result"].HostDirectory
}

func (directories *Directories) Cache() string {
	return directories.mounts["cache"].HostDirectory
}

func (directories *Directories) Focker() string {
	return directories.mounts["focker"].HostDirectory
}

func (directories *Directories) Staging() string {
	return directories.mounts["staging"].HostDirectory
}

func (directories *Directories) App() string {
	return directories.app
}

func (directories *Directories) Tmp() string {
	return directories.mounts["tmp"].HostDirectory
}

func (directories *Directories) Mounts() map[string]string {
	mappedDirectories := make(map[string]string)

	for _, directory := range directories.mounts {
		if directory.isMapped() {
			mappedDirectories[directory.HostDirectory] = directory.ContainerDirectory
		}
	}

	return mappedDirectories
}

func (directories *Directories) HostDirectories() []string {
	dirs := []string{}

	for _, directory := range directories.mounts {
		dirs = append(dirs, directory.HostDirectory)
	}

	return dirs
}

func (directories *Directories) HostDirectoriesToClean() []string {
	dirs := []string{
		directories.Droplet(),
		directories.Result(),
		directories.Staging(),
	}

	return dirs
}

func (d *Directory) isMapped() bool {
	return d.ContainerDirectory != ""
}
