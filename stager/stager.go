package stager

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/utils"

	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"
	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/runtime-schema/models"
)

type BuildpackRunner interface {
	Run() error
}

func RunBuildpack(writer io.Writer, runner BuildpackRunner) error {
	fmt.Fprintln(writer, "Running Buildpacks...")
	return runner.Run()
}

func NewBuildpackRunner(buildpackDir string) *buildpackrunner.Runner {
	prepareMd5BuildpacksDir(buildpackDir, "/tmp/buildpacks")
	var err error
	dirs := []string{}
	if dirs, err = utils.SubDirs(buildpackDir); err != nil {
		log.Fatalf(" %s", err)
	}
	config := models.NewCircusTailorConfig(dirs)
	return buildpackrunner.New(&config)
}

func ValidateStagedApp(directories *config.Directories) error {
	if _, err := os.Stat(directories.Droplet() + "/app"); err != nil {
		return fmt.Errorf("Staging failed - have you added a buildpack for this type of application?")
	}
	if _, err := os.Stat(directories.Droplet() + "/staging_info.yml"); err != nil {
		return fmt.Errorf("Staging failed - no staging info was produced by the matching buildpack!")
	}
	return nil
}

func prepareMd5BuildpacksDir(src string, dst string) {
	os.MkdirAll(src, 0755)
	os.MkdirAll(dst, 0755)
	var err error
	dirs := []string{}
	if dirs, err = utils.SubDirs(src); err != nil {
		log.Fatalf(" %s", err)
	}
	for _, dir := range dirs {
		if err := os.Symlink(src+"/"+dir, dst+"/"+md5sum(dir)); err != nil {
			log.Fatalf(" %s", err)
		}
	}
}

func md5sum(src string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(src)))
}
