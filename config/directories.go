package config

type Directories struct {
	Mounts map[string]string
}

func NewDirectories(cloudFockerHomeDir string) *Directories {
	directories := &Directories{
		Mounts: map[string]string{
			"buildpacks": cloudFockerHomeDir + "/buildpacks",
			"droplet":    cloudFockerHomeDir + "/droplet",
		},
	}
	return directories
}

func (directories *Directories) Buildpacks() string {
	return directories.Mounts["buildpacks"]
}

func (directories *Directories) Droplet() string {
	return directories.Mounts["droplet"]
}
