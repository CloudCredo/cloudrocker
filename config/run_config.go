package config

import (
	"github.com/hatofmonkeys/cloudfocker/utils"
)

type RunConfig struct {
	ContainerName  string
	ImageTag       string
	PublishedPorts map[int]int
	Mounts         map[string]string
	RunCommand	[]string
	Daemon	bool
}

func NewStageRunConfig() (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName: "cloudfocker-staging",
		ImageTag: "cloudfocker-base:latest",
		Mounts: map[string]string{
			utils.Cloudfockerhome()+"/droplet": "/tmp/droplet",
			utils.Cloudfockerhome()+"/result": "/tmp/result",
			utils.Cloudfockerhome()+"/buildpacks": "/tmp/cloudfockerbuildpacks",
			utils.Cloudfockerhome()+"/cache": "/tmp/cache",
			utils.Cloudfockerhome()+"/focker": "/fock",
		},
		RunCommand: []string{"/focker/fock", "stage"},
	}
	return
}