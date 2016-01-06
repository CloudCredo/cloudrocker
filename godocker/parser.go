package godocker

import (
	"io/ioutil"
	"log"
	"os/user"
	"sort"
	"strconv"
	"strings"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cloudcredo/cloudrocker/config"
)

func ParseCreateContainerOptions(config *config.ContainerConfig) docker.CreateContainerOptions {
	var options = docker.CreateContainerOptions{
		Name: config.ContainerName,
		Config: &docker.Config{
			User:  userID(),
			Env:   parseEnvVars(config.EnvVars),
			Image: config.SrcImageTag,
			Cmd:   config.Command,
		},
		HostConfig: &docker.HostConfig{
			Binds:        parseBinds(config.Mounts),
			PortBindings: parsePublishedPorts(config.PublishedPorts),
		},
	}
	return options
}

func WriteRuntimeDockerfile(config *config.ContainerConfig) {
	var dockerfile string

	dockerfile = runtimeInitialDockerfileString()
	dockerfile = dockerfile + envVarDockerfileString(config.EnvVars)
	dockerfile = dockerfile + commandDockerfileString(config.Command)

	ioutil.WriteFile(config.DropletDir+"/Dockerfile", []byte(dockerfile), 0644)
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

func parsePublishedPorts(publishedPorts map[int]int) map[docker.Port][]docker.PortBinding {
	var parsedPublishedPorts = make(map[docker.Port][]docker.PortBinding)
	for hostPort, containerPort := range publishedPorts {
		parsedPublishedPorts[docker.Port(strconv.Itoa(hostPort)+"/tcp")] = []docker.PortBinding{
			{
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

func runtimeInitialDockerfileString() string {
	return `FROM cloudrocker-base:latest
COPY droplet.tgz /app/
RUN chown vcap:vcap /app && cd /app && su vcap -c "tar zxf droplet.tgz" && rm droplet.tgz
EXPOSE 8080
USER vcap
WORKDIR /app
`
}

func baseImageDockerfileString(srcImageTag string) string {
	return `FROM ` + srcImageTag + `
RUN id vcap || /usr/sbin/useradd -mU -u ` + userID() + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`
}

func envVarDockerfileString(envVars map[string]string) string {
	var envVarStrings []string
	for envVarKey, envVarVal := range envVars {
		if envVarVal != "" {
			envVarStrings = append(envVarStrings, "ENV "+envVarKey+" "+envVarVal+"\n")
		}
	}
	sort.Strings(envVarStrings)
	return strings.Join(envVarStrings, "")
}

func commandDockerfileString(command []string) string {
	for index, commandElement := range command {
		command[index] = strings.Replace(commandElement, `"`, `\"`, -1)
	}
	commandString := `CMD ["`
	commandString = commandString + strings.Join(command, `", "`)
	commandString = commandString + "\"]\n"
	return commandString
}
