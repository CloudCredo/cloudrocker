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

func ParseRunCommand(config *config.ContainerConfig) (runCmd []string) {
	runCmd = append(runCmd, userString())
	runCmd = append(runCmd, parseContainerName(config.ContainerName)...)
	runCmd = append(runCmd, parseDaemon(config.Daemon)...)
	runCmd = append(runCmd, parseMounts(config.Mounts)...)
	runCmd = append(runCmd, parsePublishedPorts(config.PublishedPorts)...)
	runCmd = append(runCmd, parseEnvVars(config.EnvVars)...)
	runCmd = append(runCmd, parseSrcImageTag(config.SrcImageTag)...)
	runCmd = append(runCmd, parseCommand(config.Command)...)
	return
}

func ParseCreateContainer(config *config.ContainerConfig) docker.CreateContainerOptions {
	var options = docker.CreateContainerOptions{
		Name: config.ContainerName,
		Config: &docker.Config{
			User:  userID(),
			Env:   parseEnvVarsOptions(config.EnvVars),
			Image: config.SrcImageTag,
			Cmd:   config.Command,
		},
		HostConfig: &docker.HostConfig{
			Binds:        parseBinds(config.Mounts),
			PortBindings: parsePublishedPortsOptions(config.PublishedPorts),
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

func userString() string {
	return "-u=" + userID()
}

func userID() string {
	var thisUser *user.User
	var err error
	if thisUser, err = user.Current(); err != nil {
		log.Fatalf(" %s", err)
	}
	return thisUser.Uid
}

func parseContainerName(containerName string) (parsedContainerName []string) {
	parsedContainerName = append(parsedContainerName, "--name="+containerName)
	return
}

func parseDaemon(daemon bool) (parsedDaemon []string) {
	var daemonString string
	if daemon {
		daemonString = "-d"
		parsedDaemon = append(parsedDaemon, daemonString)
	}
	return
}

func parseMounts(mounts map[string]string) (parsedMounts []string) {
	for src, dst := range mounts {
		parsedMounts = append(parsedMounts,
			"--volume="+src+":"+dst)
	}
	sort.Strings(parsedMounts)
	return
}

func parseBinds(mounts map[string]string) (parsedBinds []string) {
	for src, dst := range mounts {
		parsedBinds = append(parsedBinds, src+":"+dst)
	}
	sort.Strings(parsedBinds)
	return
}

func parsePublishedPorts(publishedPorts map[int]int) (parsedPublishedPorts []string) {
	for host, cont := range publishedPorts {
		parsedPublishedPorts = append(parsedPublishedPorts,
			"--publish="+strconv.Itoa(host)+":"+strconv.Itoa(cont))
	}
	sort.Strings(parsedPublishedPorts)
	return
}

func parsePublishedPortsOptions(publishedPorts map[int]int) map[docker.Port][]docker.PortBinding {
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
			parsedEnvVars = append(parsedEnvVars,
				"--env=\""+key+"="+val+"\"")
		}
	}
	sort.Strings(parsedEnvVars)
	return
}

func parseEnvVarsOptions(envVars map[string]string) (parsedEnvVars []string) {
	for key, val := range envVars {
		if val != "" {
			parsedEnvVars = append(parsedEnvVars, key+"="+val)
		}
	}
	sort.Strings(parsedEnvVars)
	return
}

func parseSrcImageTag(imageTag string) (parsedSrcImageTag []string) {
	parsedSrcImageTag = append(parsedSrcImageTag, imageTag)
	return
}

func parseCommand(command []string) (parsedCommand []string) {
	parsedCommand = append(parsedCommand, command...)
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
