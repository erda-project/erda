package pvolumes

import (
	"github.com/erda-project/erda/apistructs"
)

func GenerateTaskDiceFileVolume(fileName, fileUUID, fileContainerPath string) apistructs.MetadataField {
	vo := apistructs.MetadataField{
		Name:  fileName,
		Value: fileContainerPath,
		Labels: map[string]string{
			VoLabelKeyDiceFileUUID: fileUUID,
		},
	}
	return vo
}
