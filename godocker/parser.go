package docker

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
			User:         userID(),
			Env:          parseEnvVars(config.EnvVars),
			Image:        config.SrcImageTag,
			Cmd:          config.Command,
			Mounts:       parseMounts(config.Mounts),
			AttachStdout: parseDaemon(config.Daemon),
			AttachStderr: parseDaemon(config.Daemon),
			ExposedPorts: parseExposedPorts(config.PublishedPorts),
		},
		HostConfig: &docker.HostConfig{
			Binds:        parseBinds(config.Mounts),
			PortBindings: parsePublishedPorts(config.PublishedPorts),
			NetworkMode:  "bridge",
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

type ByHostPath []docker.Mount

func (slice ByHostPath) Len() int           { return len(slice) }
func (slice ByHostPath) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }
func (slice ByHostPath) Less(i, j int) bool { return slice[i].Source < slice[j].Source }

func parseMounts(mounts map[string]string) (parsedMounts []docker.Mount) {
	for hostPath, containerPath := range mounts {
		parsedMounts = append(parsedMounts, docker.Mount{
			Source:      hostPath,
			Destination: containerPath,
			RW:          true,
		})
	}
	sort.Sort(ByHostPath(parsedMounts))
	return
}

func parseBinds(mounts map[string]string) (parsedBinds []string) {
	for hostPath, containerPath := range mounts {
		parsedBinds = append(parsedBinds, hostPath+":"+containerPath)
	}
	sort.Strings(parsedBinds)
	return
}

func parseExposedPorts(publishedPorts map[int]int) map[docker.Port]struct{} {
	var parsedExposedPorts = make(map[docker.Port]struct{})
	for hostPort := range publishedPorts {
		parsedExposedPorts[convertHostPort(hostPort)] = struct{}{}
	}
	return parsedExposedPorts
}

func parsePublishedPorts(publishedPorts map[int]int) map[docker.Port][]docker.PortBinding {
	var parsedPublishedPorts = make(map[docker.Port][]docker.PortBinding)
	for hostPort, containerPort := range publishedPorts {
		parsedPublishedPorts[convertHostPort(hostPort)] = []docker.PortBinding{
			{
				HostPort: strconv.Itoa(containerPort),
			},
		}
	}
	return parsedPublishedPorts
}

func convertHostPort(hostPort int) docker.Port {
	return docker.Port(strconv.Itoa(hostPort) + "/tcp")
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

func parseDaemon(daemon bool) bool {
	return !daemon
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
