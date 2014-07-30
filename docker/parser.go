package docker

import (
	"log"
	"os/user"
	"sort"
	"strconv"

	"github.com/hatofmonkeys/cloudfocker/config"
)

func ParseRunCommand(config *config.RunConfig) (runCmd []string) {
	mounts := parseMounts(config.Mounts)
	sort.Strings(mounts)
	runCmd = append(runCmd, mounts...)
	runCmd = append(runCmd, userString())
	runCmd = append(runCmd, parseContainerName(config.ContainerName)...)
	runCmd = append(runCmd, parseDaemon(config.Daemon)...)
	envVars := parseEnvVars(config.EnvVars)
	sort.Strings(envVars)
	runCmd = append(runCmd, envVars...)
	publishedPorts := parsePublishedPorts(config.PublishedPorts)
	sort.Strings(publishedPorts)
	runCmd = append(runCmd, publishedPorts...)
	runCmd = append(runCmd, parseImageTag(config.ImageTag)...)
	runCmd = append(runCmd, parseCommand(config.Command)...)
	return
}

func parseMounts(mounts map[string]string) (parsedMounts []string) {
	for src, dst := range mounts {
		parsedMounts = append(parsedMounts,
			"--volume="+src+":"+dst)
	}
	return
}

func userString() string {
	var thisUser *user.User
	var err error
	if thisUser, err = user.Current(); err != nil {
		log.Fatalf(" %s", err)
	}
	return "-u=" + thisUser.Uid
}

func parseContainerName(containerName string) (parsedContainerName []string) {
	parsedContainerName = append(parsedContainerName, "--name="+containerName)
	return
}

func parseImageTag(imageTag string) (parsedImageTag []string) {
	parsedImageTag = append(parsedImageTag, imageTag)
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

func parseEnvVars(envVars map[string]string) (parsedEnvVars []string) {
	for key, val := range envVars {
		parsedEnvVars = append(parsedEnvVars,
			"--env=\""+key+"="+val+"\"")
	}
	return
}

func parsePublishedPorts(publishedPorts map[int]int) (parsedEnvVars []string) {
	for host, cont := range publishedPorts {
		parsedEnvVars = append(parsedEnvVars,
			"--publish="+strconv.Itoa(host)+":"+strconv.Itoa(cont))
	}
	return
}

func parseCommand(command []string) (parsedCommand []string) {
	parsedCommand = append(parsedCommand, command...)
	return
}
