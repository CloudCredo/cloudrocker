package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const launcher = `
cd "$1"

if [ -d .profile.d ]; then
  for env_file in .profile.d/*; do
    source $env_file
  done
fi

shift

eval "$@"
`

func GetRootfsUrl() string {
	url := os.Getenv("ROCKER_ROOTFS_URL")
	if url == "" {
		url = "https://s3.amazonaws.com/blob.cfblob.com/fee97b71-17d7-4fab-a5b0-69d4112521e6"
	}
	return url
}

func CloudrockerHome() string {
	cfhome := os.Getenv("CLOUDROCKER_HOME")
	if cfhome == "" {
		cfhome = os.Getenv("HOME") + "/.cloudrocker"
	}
	return cfhome
}

func SubDirs(dir string) ([]string, error) {
	var contents []os.FileInfo
	var err error
	dirs := []string{}
	if contents, err = ioutil.ReadDir(dir); err != nil {
		return dirs, err
	}
	for _, file := range contents {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func CopyRockerBinaryToDir(destinationDir string) error {
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return err
	}
	var rockPath string
	var err error
	if rockPath, err = exec.LookPath("rock"); err != nil {
		return fmt.Errorf("Could not find rock binary, please install it in your path")
	}
	newRockPath := destinationDir + "/rock"
	if err := Cp(rockPath, newRockPath); err != nil {
		return err
	}
	if err := os.Chmod(newRockPath, 0755); err != nil {
		return err
	}
	return nil
}

func AddLauncherRunScript(appDir string) error {
	return ioutil.WriteFile(appDir+"/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", []byte(launcher), 0644)
}

func Pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %s", err)
	}
	return pwd
}

//C&P(ha!) from https://gist.github.com/elazarl/5507969
//src and dest swapped for sanity
func Cp(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
