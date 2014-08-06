package factories

import (
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/nu7hatch/gouuid"
)

func GenerateGuid() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic("Failed to generate a GUID.  Craziness.")
	}

	return guid.String()
}

func BuildTaskWithRunAction(stack string, memoryMB int, diskMB int, path string, args []string) models.Task {
	return models.Task{
		Guid:     GenerateGuid(),
		MemoryMB: memoryMB,
		DiskMB:   diskMB,
		Actions: []models.ExecutorAction{
			{Action: models.RunAction{
				Path: path,
				Args: args,
			}},
		},
		Stack: stack,
	}
}
