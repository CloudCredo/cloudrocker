package stager

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
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
	config := models.NewCircusTailorConfig(subDirs(buildpackDir))
	return buildpackrunner.New(&config)
}

func prepareMd5BuildpacksDir(src string, dst string) {
	os.MkdirAll(src, 0755)
	os.MkdirAll(dst, 0755)
	dirs := subDirs(src)
	for _, dir := range dirs {
		if err := os.Symlink(src+"/"+dir, dst+"/"+md5sum(dir)); err != nil {
			log.Fatalf(" %s", err)
		}
	}
}

func subDirs(dir string) (dirs []string) {
	var contents []os.FileInfo
	var err error
	if contents, err = ioutil.ReadDir(dir); err != nil {
		log.Fatalf(" %s", err)
	}
	for _, file := range contents {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return
}

func md5sum(src string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(src)))
}
