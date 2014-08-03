package buildpack

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/hatofmonkeys/cloudfocker/utils"
)

func Add(writer io.Writer, url string, buildpackDir string) {
	err := os.MkdirAll(buildpackDir, 0755)
	if err != nil {
		log.Fatalf("Buildpack directory creation error: %s", err)
	}
	fmt.Fprintln(writer, "Downloading buildpack...")
	cmd := exec.Command("git", "clone", "--depth=1", "--recursive", url)
	cmd.Stdout = writer
	cmd.Stderr = writer
	cmd.Dir = buildpackDir
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error downloading buildpack: %s", err)
	}
	fmt.Fprintln(writer, "Downloaded buildpack.")
}

func Delete(writer io.Writer, buildpack string, buildpackDir string) error {
	if err := os.RemoveAll(buildpackDir + "/" + buildpack); err != nil {
		return err
	}
	fmt.Fprintln(writer, "Deleted buildpack.")
	return nil
}

func List(writer io.Writer, buildpackDir string) (err error) {
	if buildpacks, err := utils.SubDirs(buildpackDir); err == nil {
		for _, buildpack := range buildpacks {
			fmt.Fprintln(writer, buildpack)
		}
		if len(buildpacks) == 0 {
			fmt.Fprintln(writer, "No buildpacks installed")
		}
	}
	return err
}

func AtLeastOneBuildpackIn(dir string) error {
	var subDirs []string
	var err error
	if subDirs, err = utils.SubDirs(dir); err != nil {
		return err
	}
	if len(subDirs) == 0 {
		return fmt.Errorf("No buildpacks detected - please add one")
	}
	return nil
}
