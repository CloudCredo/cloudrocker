package models

import (
	"crypto/md5"
	"flag"
	"fmt"
	"path"
	"strings"
)

type CircusTailorConfig struct {
	*flag.FlagSet

	values map[string]*string

	buildpacksDir  *string
	appDir         *string
	ExecutablePath string

	buildArtifactsCacheDir *string
	outputDropletDir       *string
	outputMetadataDir      *string
	buildpackOrder         *string
}

const (
	circusTailorAppDirFlag                 = "appDir"
	circusTailorOutputDropletDirFlag       = "outputDropletDir"
	circusTailorOutputMetadataDirFlag      = "outputMetadataDir"
	circusTailorBuildpacksDirFlag          = "buildpacksDir"
	circusTailorBuildArtifactsCacheDirFlag = "buildArtifactsCacheDir"
	circusTailorBuildpackOrderFlag         = "buildpackOrder"
)

var circusTailorDefaults = map[string]string{
	circusTailorAppDirFlag:                 "/app",
	circusTailorOutputDropletDirFlag:       "/tmp/droplet",
	circusTailorOutputMetadataDirFlag:      "/tmp/result",
	circusTailorBuildpacksDirFlag:          "/tmp/buildpacks",
	circusTailorBuildArtifactsCacheDirFlag: "/tmp/cache",
}

func NewCircusTailorConfig(buildpacks []string) CircusTailorConfig {
	flagSet := flag.NewFlagSet("tailor", flag.ExitOnError)

	appDir := flagSet.String(
		circusTailorAppDirFlag,
		circusTailorDefaults[circusTailorAppDirFlag],
		"directory containing raw app bits",
	)

	outputDropletDir := flagSet.String(
		circusTailorOutputDropletDirFlag,
		circusTailorDefaults[circusTailorOutputDropletDirFlag],
		"directory in which to write the droplet",
	)

	outputMetadataDir := flagSet.String(
		circusTailorOutputMetadataDirFlag,
		circusTailorDefaults[circusTailorOutputMetadataDirFlag],
		"directory in which to write the app metadata",
	)

	buildpacksDir := flagSet.String(
		circusTailorBuildpacksDirFlag,
		circusTailorDefaults[circusTailorBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	buildArtifactsCacheDir := flagSet.String(
		circusTailorBuildArtifactsCacheDirFlag,
		circusTailorDefaults[circusTailorBuildArtifactsCacheDirFlag],
		"directory to store cached artifacts to buildpacks",
	)

	buildpackOrder := flagSet.String(
		circusTailorBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	return CircusTailorConfig{
		FlagSet: flagSet,

		ExecutablePath:         "/tmp/circus/tailor",
		appDir:                 appDir,
		outputDropletDir:       outputDropletDir,
		outputMetadataDir:      outputMetadataDir,
		buildpacksDir:          buildpacksDir,
		buildArtifactsCacheDir: buildArtifactsCacheDir,
		buildpackOrder:         buildpackOrder,

		values: map[string]*string{
			circusTailorAppDirFlag:                 appDir,
			circusTailorOutputDropletDirFlag:       outputDropletDir,
			circusTailorOutputMetadataDirFlag:      outputMetadataDir,
			circusTailorBuildpacksDirFlag:          buildpacksDir,
			circusTailorBuildArtifactsCacheDirFlag: buildArtifactsCacheDir,
			circusTailorBuildpackOrderFlag:         buildpackOrder,
		},
	}
}

func (s CircusTailorConfig) Path() string {
	return s.ExecutablePath
}

func (s CircusTailorConfig) Args() []string {
	argv := []string{}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s=%s", flag.Name, *s.values[flag.Name]))
	})

	return argv
}

func (s CircusTailorConfig) Validate() error {
	var missingFlags []string

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		schemaFlag, ok := s.values[flag.Name]
		if !ok {
			return
		}

		value := *schemaFlag
		if value == "" {
			missingFlags = append(missingFlags, "-"+flag.Name)
		}
	})

	if len(missingFlags) > 0 {
		return fmt.Errorf("missing flags: %s", strings.Join(missingFlags, ", "))
	}

	return nil
}

func (s CircusTailorConfig) AppDir() string {
	return *s.appDir
}

func (s CircusTailorConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.BuildpacksDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
}

func (s CircusTailorConfig) BuildpackOrder() []string {
	return strings.Split(*s.buildpackOrder, ",")
}

func (s CircusTailorConfig) BuildpacksDir() string {
	return *s.buildpacksDir
}

func (s CircusTailorConfig) BuildArtifactsCacheDir() string {
	return *s.buildArtifactsCacheDir
}

func (s CircusTailorConfig) OutputDropletDir() string {
	return *s.outputDropletDir
}

func (s CircusTailorConfig) OutputMetadataDir() string {
	return *s.outputMetadataDir
}

func (s CircusTailorConfig) OutputMetadataPath() string {
	return path.Join(s.OutputMetadataDir(), "result.json")
}
