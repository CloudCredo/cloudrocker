package config

import (
	"github.com/hatofmonkeys/cloudfocker/utils"
)

type RunConfig struct {
	ContainerName  string
	ImageTag       string
	PublishedPorts map[int]int
	Mounts         map[string]string
	Command     []string
	Daemon         bool
}

func NewStageRunConfig(cloudfoundryAppDir string) (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName: "cloudfocker-staging",
		ImageTag:      "cloudfocker-base:latest",
		Mounts: map[string]string{
			cloudfoundryAppDir:                      "/app",
			utils.Cloudfockerhome() + "/droplet":    "/tmp/droplet",
			utils.Cloudfockerhome() + "/result":     "/tmp/result",
			utils.Cloudfockerhome() + "/buildpacks": "/tmp/cloudfockerbuildpacks",
			utils.Cloudfockerhome() + "/cache":      "/tmp/cache",
			utils.Cloudfockerhome() + "/focker":     "/focker",
		},
		Command: []string{"/focker/fock", "stage"},
	}
	return
}
