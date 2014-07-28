package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func GetRootfsUrl() string {
	url := os.Getenv("FOCKER_ROOTFS_URL")
	if url == "" {
		url = "https://s3.amazonaws.com/blob.cfblob.com/fee97b71-17d7-4fab-a5b0-69d4112521e6"
	}
	return url
}

func Cloudfockerhome() string {
	cfhome := os.Getenv("CLOUDFOCKER_HOME")
	if cfhome == "" {
		cfhome = os.Getenv("HOME") + "/.cloudfocker"
	}
	return cfhome
}

func CreateAndCleanAppDirs(cloudfockerhomeDir string) error {
	dirs := map[string]bool{"/buildpacks": false, "/droplet": true, "/cache": false, "/result": true}
	for dir, clean := range dirs {
		if clean {
			if err := os.RemoveAll(cloudfockerhomeDir + dir); err != nil {
				return err
			}
		}
	}
	for dir, _ := range dirs {
		if err := os.MkdirAll(cloudfockerhomeDir+dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func AtLeastOneBuildpackIn(dir string) error {
	var subDirs []string
	var err error
	if subDirs, err = SubDirs(dir); err != nil {
		return err
	}
	if len(subDirs) == 0 {
		return fmt.Errorf("No buildpacks detected - please add one")
	}
	return nil
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

func CopyFockerBinaryToOwnDir(cloudfockerhome string) error {
	if err := os.MkdirAll(cloudfockerhome+"/focker", 0755); err != nil {
		return err
	}
	var fockPath string
	var err error
	if fockPath, err = exec.LookPath("fock"); err != nil {
		return fmt.Errorf("Could not find fock binary, please install it in your path")
	}
	newFockPath := cloudfockerhome + "/focker/fock"
	if err := Cp(fockPath, newFockPath); err != nil {
		return err
	}
	if err := os.Chmod(newFockPath, 0755); err != nil {
		return err
	}
	return nil
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
