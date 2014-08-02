package config

import (
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/hatofmonkeys/cloudfocker/utils"
)

type RunConfig struct {
	ContainerName  string
	ImageTag       string
	PublishedPorts map[int]int
	Mounts         map[string]string
	Command        []string
	Daemon         bool
	EnvVars        map[string]string
}

func NewStageRunConfig(cloudfoundryAppDir string) (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName: "cloudfocker-staging",
		ImageTag:      "cloudfocker-base:latest",
		Mounts: map[string]string{
			cloudfoundryAppDir:                      "/app",
			utils.CloudfockerHome() + "/droplet":    "/tmp/droplet",
			utils.CloudfockerHome() + "/result":     "/tmp/result",
			utils.CloudfockerHome() + "/buildpacks": "/tmp/cloudfockerbuildpacks",
			utils.CloudfockerHome() + "/cache":      "/tmp/cache",
			utils.CloudfockerHome() + "/focker":     "/focker",
		},
		Command: []string{"/focker/fock", "stage-internal"},
	}
	return
}

func NewRuntimeRunConfig(cloudfoundryDropletDir string) (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName:  "cloudfocker-runtime",
		ImageTag:       "cloudfocker-base:latest",
		PublishedPorts: map[int]int{8080: 8080},
		Mounts: map[string]string{
			cloudfoundryDropletDir + "/app": "/app",
		},
		Command: append([]string{"/bin/bash", "/app/cloudfocker-start.sh", "/app"},
			parseStartCommand(cloudfoundryDropletDir)...),
		Daemon: true,
		EnvVars: map[string]string{
			"HOME":   "/app",
			"TMPDIR": "/app/tmp",
			"PORT":   "8080",
		},
	}
	return
}

type StagingInfoYml struct {
	DetectedBuildpack string `yaml:"detected_buildpack"`
	StartCommand      string `yaml:"start_command"`
}

func parseStartCommand(cloudfoundryDropletDir string) []string {
	file, err := os.Open(cloudfoundryDropletDir + "/staging_info.yml")
	if err != nil {
		log.Fatalf("File does not exist: %s", err)
	}
	stagingInfo := new(StagingInfoYml)
	decoder := candiedyaml.NewDecoder(file)
	err = decoder.Decode(stagingInfo)
	if err != nil {
		log.Fatalf("Failed to decode document: %s", err)
	}
	return strings.Split(stagingInfo.StartCommand, " ")
}
