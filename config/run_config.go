package config

import (
	"log"
	"os"
	"strings"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/hatofmonkeys/cloudfocker/utils"
)

type RunConfig struct {
	ContainerName  string
	Daemon         bool
	Mounts         map[string]string
	PublishedPorts map[int]int
	EnvVars        map[string]string
	ImageTag       string
	Command        []string
}

func NewStageRunConfig(cloudfoundryAppDir string) (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName: "cloudfocker-staging",
		Mounts: map[string]string{
			cloudfoundryAppDir:                      "/app",
			utils.CloudfockerHome() + "/droplet":    "/tmp/droplet",
			utils.CloudfockerHome() + "/result":     "/tmp/result",
			utils.CloudfockerHome() + "/buildpacks": "/tmp/cloudfockerbuildpacks",
			utils.CloudfockerHome() + "/cache":      "/tmp/cache",
			utils.CloudfockerHome() + "/focker":     "/focker",
		},
		ImageTag: "cloudfocker-base:latest",
		Command:  []string{"/focker/fock", "stage-internal"},
	}
	return
}

func NewRuntimeRunConfig(cloudfoundryDropletDir string) (runConfig *RunConfig) {
	runConfig = &RunConfig{
		ContainerName: "cloudfocker-runtime",
		Daemon:        true,
		Mounts: map[string]string{
			cloudfoundryDropletDir + "/app": "/app",
		},
		PublishedPorts: map[int]int{8080: 8080},
		EnvVars: map[string]string{
			"HOME":   "/app",
			"TMPDIR": "/app/tmp",
			"PORT":   "8080",
			"VCAP_SERVICES": vcapServices(cloudfoundryDropletDir),
		},
		ImageTag: "cloudfocker-base:latest",
		Command: append([]string{"/bin/bash", "/app/cloudfocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", "/app"},
			parseStartCommand(cloudfoundryDropletDir)...),
	}
	return
}

func vcapServices(cloudfoundryDropletDir string) (services string) {
	servicesBytes, err := ioutil.ReadFile(cloudfoundryDropletDir + "/app/vcap_services.json")
	if err != nil {
		return
	}
	services = string(servicesBytes)
	return
}

type StagingInfoYml struct {
	DetectedBuildpack string `yaml:"detected_buildpack"`
	StartCommand      string `yaml:"start_command"`
}

type ProcfileYml struct {
	Web string `yaml:"web"`
}

func parseStartCommand(cloudfoundryDropletDir string) (startCommand []string) {
	stagingInfoFile, err := os.Open(cloudfoundryDropletDir + "/staging_info.yml")
	if err == nil {
		stagingInfo := new(StagingInfoYml)
		decoder := candiedyaml.NewDecoder(stagingInfoFile)
		err = decoder.Decode(stagingInfo)
		if err != nil {
			log.Fatalf("Failed to decode document: %s", err)
		}
		startCommand = strings.Split(stagingInfo.StartCommand, " ")
		if startCommand[0] != "" {
			return
		}
		procfileFile, err := os.Open(cloudfoundryDropletDir + "/app/Procfile")
		if err == nil {
			procfileInfo := new(ProcfileYml)
			decoder := candiedyaml.NewDecoder(procfileFile)
			err = decoder.Decode(procfileInfo)
			if err != nil {
				log.Fatalf("Failed to decode document: %s", err)
			}
			startCommand = strings.Split(procfileInfo.Web, " ")
			return
		}
	}
	log.Fatal("Unable to find staging_info.yml")
	return
}
