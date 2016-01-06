package godocker

import (
	"io/ioutil"
	"log"
	"os/user"
	"sort"
	"strconv"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cloudcredo/cloudrocker/config"
)

func ParseCreateContainerOptions(config *config.ContainerConfig) docker.CreateContainerOptions {
	var options = docker.CreateContainerOptions{
		Name: config.ContainerName,
		Config: &docker.Config{
			User:    userID(),
			Env:     parseEnvVars(config.EnvVars),
			Image:   config.SrcImageTag,
			Cmd:     config.Command,
			Volumes: parseVolumes(config.Mounts),
		},
		HostConfig: &docker.HostConfig{
			Binds:        parseBinds(config.Mounts),
			PortBindings: parsePublishedPorts(config.PublishedPorts),
		},
	}
	return options
}

func WriteBaseImageDockerfile(config *config.ContainerConfig) {
	var dockerfile string

	dockerfile = baseImageDockerfileString(config.SrcImageTag)

	ioutil.WriteFile(config.BaseConfigDir+"/Dockerfile", []byte(dockerfile), 0644)
}

func userID() string {
	var thisUser *user.User
	var err error
	if thisUser, err = user.Current(); err != nil {
		log.Fatalf(" %s", err)
	}
	return thisUser.Uid
}

func parseBinds(mounts map[string]string) (parsedBinds []string) {
	for hostPath, containerPath := range mounts {
		parsedBinds = append(parsedBinds, hostPath+":"+containerPath)
	}
	sort.Strings(parsedBinds)
	return
}

func parseVolumes(mounts map[string]string) map[string]struct{} {
	var parsedVolumes = make(map[string]struct{})
	for _, containerPath := range mounts {
		parsedVolumes[containerPath] = struct{}{}
	}
	return parsedVolumes
}

func parsePublishedPorts(publishedPorts map[int]int) map[docker.Port][]docker.PortBinding {
	var parsedPublishedPorts = make(map[docker.Port][]docker.PortBinding)
	for hostPort, containerPort := range publishedPorts {
		parsedPublishedPorts[docker.Port(strconv.Itoa(hostPort)+"/tcp")] = []docker.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(containerPort),
			},
		}
	}
	return parsedPublishedPorts
}

func parseEnvVars(envVars map[string]string) (parsedEnvVars []string) {
	for key, val := range envVars {
		if val != "" {
			parsedEnvVars = append(parsedEnvVars, key+"="+val)
		}
	}
	sort.Strings(parsedEnvVars)
	return
}

func baseImageDockerfileString(srcImageTag string) string {
	return `FROM ` + srcImageTag + `
RUN id vcap || /usr/sbin/useradd -mU -u ` + userID() + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`
}
