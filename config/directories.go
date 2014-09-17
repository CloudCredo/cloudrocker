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
