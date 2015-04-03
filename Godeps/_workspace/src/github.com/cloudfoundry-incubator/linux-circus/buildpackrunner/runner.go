package buildpackrunner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/linux-circus/protocol"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/runtime-schema/models"

	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/candiedyaml"
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
		return newDescriptiveError(err, "Failed to set up filesystem when generating droplet")
	}

	//detect, compile, release
	detectedBuildpack, detectedBuildpackDir, detectOutput, err := runner.detect()
	if err != nil {
		return err
	}

	err = runner.compile(detectedBuildpackDir)
	if err != nil {
		return newDescriptiveError(err, "Failed to compile droplet")
	}

	startCommand, err := runner.detectStartCommandFromProcfile()
	if err != nil {
		return newDescriptiveError(err, "Failed to read command from Procfile")
	}

	releaseInfo, err := runner.release(detectedBuildpackDir, startCommand)
	if err != nil {
		return newDescriptiveError(err, "Failed to build droplet release")
	}

	if len(releaseInfo.DefaultProcessTypes.Web) == 0 {
		printError("No start command detected; command must be provided at runtime")
	}

	//generate staging_info.yml and result json file
	err = runner.saveInfo(detectedBuildpack, detectOutput, releaseInfo)
	if err != nil {
		return newDescriptiveError(err, "Failed to encode generated metadata")
	}

	//prepare the final droplet directory
	err = runner.copyApp(runner.config.AppDir(), path.Join(runner.config.OutputDropletDir(), "app"))
	if err != nil {
		return newDescriptiveError(err, "Failed to copy compiled droplet")
	}

	err = os.MkdirAll(path.Join(runner.config.OutputDropletDir(), "tmp"), 0755)
	if err != nil {
		return newDescriptiveError(err, "Failed to set up droplet filesystem")
	}

	err = os.MkdirAll(path.Join(runner.config.OutputDropletDir(), "logs"), 0755)
	if err != nil {
		return newDescriptiveError(err, "Failed to set up droplet filesystem")
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
		return "", newDescriptiveError(nil, "Failed to read buildpack directory '%s' for buildpack '%s'", buildpackPath, buildpack)
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
			printError(err.Error())
			continue
		}

		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "detect"), runner.config.AppDir()), output)

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\n"), nil
		}
	}

	return "", "", "", newDescriptiveError(nil, "no valid buildpacks detected")
}

func (runner *Runner) detectStartCommandFromProcfile() (string, error) {
	procFile, err := os.Open(filepath.Join(runner.config.AppDir(), "Procfile"))
	if err != nil {
		if os.IsNotExist(err) {
			// Procfiles are optional
			return "", nil
		}

		return "", err
	}

	defer procFile.Close()

	processes := map[string]string{}

	err = candiedyaml.NewDecoder(procFile).Decode(&processes)
	if err != nil {
		// clobber candiedyaml's super low-level error
		return "", errors.New("invalid YAML")
	}

	return processes["web"], nil
}

func (runner *Runner) compile(buildpackDir string) error {
	return runner.run(exec.Command(path.Join(buildpackDir, "bin", "compile"), runner.config.AppDir(), runner.config.BuildArtifactsCacheDir()), os.Stdout)
}

func (runner *Runner) release(buildpackDir string, webStartCommand string) (Release, error) {
	output := new(bytes.Buffer)

	err := runner.run(exec.Command(path.Join(buildpackDir, "bin", "release"), runner.config.AppDir()), output)
	if err != nil {
		return Release{}, err
	}

	decoder := candiedyaml.NewDecoder(output)

	parsedRelease := Release{}

	err = decoder.Decode(&parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	procfileContainsWebStartCommand := webStartCommand != ""
	if procfileContainsWebStartCommand {
		parsedRelease.DefaultProcessTypes.Web = webStartCommand
	}

	return parsedRelease, nil
}

func (runner *Runner) saveInfo(buildpack string, detectOutput string, releaseInfo Release) error {
	deaInfoFile, err := os.Create(filepath.Join(runner.config.OutputDropletDir(), "staging_info.yml"))
	if err != nil {
		return err
	}
	defer deaInfoFile.Close()

	err = candiedyaml.NewEncoder(deaInfoFile).Encode(DeaStagingInfo{
		DetectedBuildpack: detectOutput,
		StartCommand:      releaseInfo.DefaultProcessTypes.Web,
	})
	if err != nil {
		return err
	}

	resultFile, err := os.Create(runner.config.OutputMetadataPath())
	if err != nil {
		return err
	}
	defer resultFile.Close()

	executionMetadata, err := json.Marshal(protocol.ExecutionMetadata{
		StartCommand: releaseInfo.DefaultProcessTypes.Web,
	})
	if err != nil {
		return err
	}

	err = json.NewEncoder(resultFile).Encode(models.StagingResult{
		BuildpackKey:         buildpack,
		DetectedBuildpack:    detectOutput,
		ExecutionMetadata:    string(executionMetadata),
		DetectedStartCommand: map[string]string{"web": releaseInfo.DefaultProcessTypes.Web},
	})
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

func printError(message string) {
	println(message)
}
