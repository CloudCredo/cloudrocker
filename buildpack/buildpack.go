package buildpack

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
