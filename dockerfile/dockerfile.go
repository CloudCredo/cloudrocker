package dockerfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Dockerfile struct {
	Commands []string
}

func NewDockerfile() (dockerfile *Dockerfile) {
	dockerfile = new(Dockerfile)
	return
}

func (dockerfile *Dockerfile) Write(writer io.Writer) {
	fmt.Fprint(writer, dockerfile.tostring())
}

func (dockerfile *Dockerfile) Persist(location string) (err error) {
	filestring := strings.Join(dockerfile.Commands, "\n")
	err = ioutil.WriteFile(location, []byte(filestring), 0644)
	return
}

func (dockerfile *Dockerfile) Create() {
	//create a slice of Cmds, run them on the Dockerfile to populate
	cmds := []func(dockerfile *Dockerfile){
		addBaseImageCmd,
		addAddAppCmd,
		addAddBuildpackCmd,
		addAddTailorCmd,
		addRunTailorCmd,
		addExposeCmd,
		addEntrypointCmd,
	}

	for _, cmd := range cmds {
		cmd(dockerfile)
	}
}

func addBaseImageCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"FROM cloudfocker-base:latest")
}

func addAddAppCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"RUN echo 'HELLO, WORLD'")
}

func addAddBuildpackCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"RUN echo 'HELLO, WORLD'")
}

func addAddTailorCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"RUN echo 'HELLO, WORLD'")
}

func addRunTailorCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"RUN echo 'HELLO, WORLD'")
}

func addExposeCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"EXPOSE 8080")
}

func addEntrypointCmd(dockerfile *Dockerfile) {
	dockerfile.Commands = append(dockerfile.Commands,
		"ENTRYPOINT [\"/bin/bash\"]")
}

func (dockerfile *Dockerfile) tostring() (filestring string) {
	filestring = strings.Join(dockerfile.Commands, "\n") + "\n"
	return
}
