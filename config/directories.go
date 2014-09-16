package config

type Directories struct {
	Mounts map[string]string
}

func NewDirectories(cloudFockerHomeDir string) *Directories {
	directories := &Directories{
		Mounts: map[string]string{
			"buildpacks": cloudFockerHomeDir + "/buildpacks",
		},
	}
	return directories
}

func (directories *Directories) Buildpacks() string {
	return directories.Mounts["buildpacks"]
}
