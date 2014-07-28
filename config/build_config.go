package config

type BuildConfig struct {
	ImageTag     string
	AddFiles     map[string]string
	StartCommand []string
}

func NewStageBuildConfig() (buildConfig *BuildConfig) {
	buildConfig = &BuildConfig{
		ImageTag: "cloudfocker-base:latest",
		AddFiles: map[string]string{
			"fock": "/",
		},
		StartCommand: []string{"/fock", "stage"},
	}
	return
}
