package config

import (
	"github.com/cloudcredo/cloudfocker/utils"
)

type Directories struct {
	mounts map[string]Directory
}

type Directory struct {
	HostDirectory      string
	ContainerDirectory string
}

func NewDirectories(cloudFockerHomeDir string) *Directories {
	directories := &Directories{
		mounts: map[string]Directory{
			"home":       Directory{cloudFockerHomeDir, ""},
			"buildpacks": Directory{cloudFockerHomeDir + "/buildpacks", "/tmp/cloudfockerbuildpacks"},
			"droplet":    Directory{cloudFockerHomeDir + "/droplet", "/tmp/droplet"},
			"result":     Directory{cloudFockerHomeDir + "/result", "/tmp/result"},
			"cache":      Directory{cloudFockerHomeDir + "/cache", "/tmp/cache"},
			"focker":     Directory{cloudFockerHomeDir + "/focker", "/focker"},
			"staging":    Directory{cloudFockerHomeDir + "/staging", ""},
			"app":        Directory{utils.Pwd(), ""},
		},
	}
	return directories
}

func (directories *Directories) Home() string {
	return directories.mounts["home"].HostDirectory
}

func (directories *Directories) Buildpacks() string {
	return directories.mounts["buildpacks"].HostDirectory
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
	return directories.mounts["app"].HostDirectory
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

func (d *Directory) isMapped() bool {
	return d.ContainerDirectory != ""
}
