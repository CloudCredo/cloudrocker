package config

type Directories struct {
	mounts map[string]string
}

func NewDirectories(cloudFockerHomeDir string) *Directories {
	directories := &Directories{
		mounts: map[string]string{
			"buildpacks": cloudFockerHomeDir + "/buildpacks",
			"droplet":    cloudFockerHomeDir + "/droplet",
			"result":     cloudFockerHomeDir + "/result",
			"cache":      cloudFockerHomeDir + "/cache",
			"focker":     cloudFockerHomeDir + "/focker",
		},
	}
	return directories
}

func (directories *Directories) Buildpacks() string {
	return directories.mounts["buildpacks"]
}

func (directories *Directories) Droplet() string {
	return directories.mounts["droplet"]
}

func (directories *Directories) Result() string {
	return directories.mounts["result"]
}

func (directories *Directories) Cache() string {
	return directories.mounts["cache"]
}

func (directories *Directories) Focker() string {
	return directories.mounts["focker"]
}
