package buildpackrunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/runtime-schema/models"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Runner struct {
	config *models.CircusTailorConfig
}

type descriptiveError struct {
	message string
	err     error
}

func (e descriptiveError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func newDescriptiveError(err error, message string, args ...interface{}) error {
	if len(args) == 0 {
		return descriptiveError{message: message, err: err}
	}
	return descriptiveError{message: fmt.Sprintf(message, args...), err: err}
}

type Release struct {
	DefaultProcessTypes struct {
		Web string `yaml:"web"`
	} `yaml:"default_process_types"`
}

func New(config *models.CircusTailorConfig) *Runner {
	return &Runner{
		config: config,
	}
}

func (runner *Runner) Run() error {
	//set up the world
	err := runner.makeDirectories()
	if err != nil {
		return newDescriptiveError(err, "failed to set up filesystem when generating droplet")
	}

	//detect, compile, release
	detectedBuildpack, detectedBuildpackDir, detectOutput, err := runner.detect()
	if err != nil {
		return err
	}

	err = runner.compile(detectedBuildpackDir)
	if err != nil {
		return newDescriptiveError(err, "failed to compile droplet")
	}

	releaseInfo, err := runner.release(detectedBuildpackDir)
	if err != nil {
		return newDescriptiveError(err, "failed to build droplet release")
	}

	//generate staging_info.yml and result json file
	err = runner.saveInfo(detectedBuildpack, detectOutput, releaseInfo)
	if err != nil {
		return newDescriptiveError(err, "failed to encode generated metadata")
	}

	//prepare the final droplet directory
	err = runner.copyApp(runner.config.AppDir(), path.Join(runner.config.OutputDropletDir(), "app"))
	if err != nil {
		return newDescriptiveError(err, "failed to copy compiled droplet")
	}

	err = os.MkdirAll(path.Join(runner.config.OutputDropletDir(), "tmp"), 0755)
	if err != nil {
		return newDescriptiveError(err, "failed to set up droplet filesystem")
	}

	err = os.MkdirAll(path.Join(runner.config.OutputDropletDir(), "logs"), 0755)
	if err != nil {
		return newDescriptiveError(err, "failed to set up droplet filesystem")
	}

	return nil
}

func (runner *Runner) makeDirectories() error {
	if err := os.MkdirAll(runner.config.OutputDropletDir(), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(runner.config.OutputMetadataDir(), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(runner.config.BuildArtifactsCacheDir(), 0755); err != nil {
		return err
	}

	return nil
}

func (runner *Runner) buildpackPath(buildpack string) (string, error) {
	buildpackPath := runner.config.BuildpackPath(buildpack)

	if runner.pathHasBinDirectory(buildpackPath) {
		return buildpackPath, nil
	}

	files, err := ioutil.ReadDir(buildpackPath)
	if err != nil {
		return "", newDescriptiveError(nil, "failed to read buildpack directory for buildpack: %s", buildpack)
	}

	if len(files) == 1 {
		nestedPath := path.Join(buildpackPath, files[0].Name())

		if runner.pathHasBinDirectory(nestedPath) {
			return nestedPath, nil
		}
	}

	return "", newDescriptiveError(nil, "malformed buildpack does not contain a /bin dir: %s", buildpack)
}

func (runner *Runner) pathHasBinDirectory(pathToTest string) bool {
	_, err := os.Stat(path.Join(pathToTest, "bin"))
	return err == nil
}

func (runner *Runner) detect() (string, string, string, error) {
	for _, buildpack := range runner.config.BuildpackOrder() {
		output := new(bytes.Buffer)

		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "detect"), runner.config.AppDir()), output)

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\n"), nil
		}
	}

	return "", "", "", newDescriptiveError(nil, "no valid buildpacks detected")
}

func (runner *Runner) compile(buildpackDir string) error {
	return runner.run(exec.Command(path.Join(buildpackDir, "bin", "compile"), runner.config.AppDir(), runner.config.BuildArtifactsCacheDir()), os.Stdout)
}

func (runner *Runner) release(buildpackDir string) (Release, error) {
	output := new(bytes.Buffer)

	err := runner.run(exec.Command(path.Join(buildpackDir, "bin", "release"), runner.config.AppDir()), output)
	if err != nil {
		return Release{}, err
	}

	decoder := candiedyaml.NewDecoder(output)

	var parsedRelease Release

	err = decoder.Decode(&parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	return parsedRelease, nil
}

func (runner *Runner) saveInfo(buildpack string, detectOutput string, releaseInfo Release) error {
	infoFile, err := os.Create(filepath.Join(runner.config.OutputDropletDir(), "staging_info.yml"))
	if err != nil {
		return err
	}

	defer infoFile.Close()

	resultFile, err := os.Create(runner.config.OutputMetadataPath())
	if err != nil {
		return err
	}

	defer resultFile.Close()

	info := models.StagingInfo{
		BuildpackKey:         buildpack,
		DetectedBuildpack:    detectOutput,
		DetectedStartCommand: releaseInfo.DefaultProcessTypes.Web,
	}

	err = candiedyaml.NewEncoder(infoFile).Encode(info)
	if err != nil {
		return err
	}

	err = json.NewEncoder(resultFile).Encode(info)
	if err != nil {
		return err
	}

	return nil
}

func (runner *Runner) copyApp(appDir, stageDir string) error {
	return runner.run(exec.Command("cp", "-a", appDir, stageDir), os.Stdout)
}

func (runner *Runner) run(cmd *exec.Cmd, output io.Writer) error {
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
