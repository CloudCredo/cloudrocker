package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/candiedyaml"
)

type ContainerConfig struct {
	ContainerName  string
	Daemon         bool
	Mounts         map[string]string
	PublishedPorts map[int]int
	EnvVars        map[string]string
	ImageTag       string
	Command        []string
	DropletDir     string
}

func NewStageContainerConfig(directories *Directories) (containerConfig *ContainerConfig) {
	containerConfig = &ContainerConfig{
		ContainerName: "cloudfocker-staging",
		Mounts:        directories.Mounts(),
		ImageTag:      "cloudfocker-base:latest",
		Command:       []string{"/focker/fock", "stage", "internal"},
	}
	return
}

func NewRuntimeContainerConfig(dropletDir string) (containerConfig *ContainerConfig) {
	containerConfig = &ContainerConfig{
		ContainerName: "cloudfocker-runtime",
		Daemon:        true,
		Mounts: map[string]string{
			dropletDir + "/app": "/app",
		},
		PublishedPorts: map[int]int{8080: 8080},
		EnvVars: map[string]string{
			"HOME":          "/app",
			"TMPDIR":        "/app/tmp",
			"PORT":          "8080",
			"VCAP_SERVICES": vcapServices(dropletDir),
			"DATABASE_URL":  databaseURL(dropletDir),
		},
		ImageTag: "cloudfocker-base:latest",
		Command: append([]string{"/bin/bash", "/app/cloudfocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", "/app"},
			parseStartCommand(dropletDir)...),
		DropletDir: dropletDir,
	}
	return
}

func vcapServices(dropletDir string) (services string) {
	servicesBytes, err := ioutil.ReadFile(dropletDir + "/app/vcap_services.json")
	if err != nil {
		return
	}
	services = string(servicesBytes)
	return
}

type database struct {
	Credentials struct {
		URI string
	}
}

func databaseURL(dropletDir string) (databaseURL string) {
	servicesBytes, err := ioutil.ReadFile(dropletDir + "/app/vcap_services.json")
	if err != nil {
		return
	}

	var services map[string][]database

	json.Unmarshal(servicesBytes, &services)

	for _, serviceDatabase := range services {
		if len(serviceDatabase) > 0 && serviceDatabase[0].Credentials.URI != "" {
			databaseURL = serviceDatabase[0].Credentials.URI
		}
	}

	return
}

type StagingInfoYml struct {
	DetectedBuildpack string `yaml:"detected_buildpack"`
	StartCommand      string `yaml:"start_command"`
}

type ProcfileYml struct {
	Web string `yaml:"web"`
}

func parseStartCommand(dropletDir string) (startCommand []string) {
	stagingInfoFile, err := os.Open(dropletDir + "/staging_info.yml")
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
		procfileFile, err := os.Open(dropletDir + "/app/Procfile")
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
