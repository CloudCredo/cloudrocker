package docker

import (
	"sort"

	"github.com/hatofmonkeys/cloudfocker/config"
)

func ParseRunCommand(config *config.RunConfig) (runCmd []string) {
	mounts := parseMounts(config.Mounts)
	sort.Strings(mounts)
	runCmd = append(runCmd, mounts...)
	runCmd = append(runCmd, parseContainerName(config.ContainerName)...)
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

func parseContainerName(containerName string) (parsedContainerName []string) {
	parsedContainerName = append(parsedContainerName, "--name="+containerName)
	return
}

func parseImageTag(imageTag string) (parsedImageTag []string) {
	parsedImageTag = append(parsedImageTag, imageTag)
	return
}

func parseCommand(command []string) (parsedCommand []string) {
	parsedCommand = append(parsedCommand, command...)
	return
}
