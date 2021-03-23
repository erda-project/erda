package pvolumes

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func GenerateTaskVolume(task spec.PipelineTask, namespace string, volumeID *string) apistructs.MetadataField {
	volumeMountPath := MakeTaskContainerVolumeMountDir(namespace)
	vo := apistructs.MetadataField{
		Name:  namespace,
		Value: volumeMountPath,
		Type:  string(spec.StoreTypeDiceVolumeNFS),
		Labels: map[string]string{
			VoLabelKeyContextPath:   volumeMountPath,
			VoLabelKeyContainerPath: MakeTaskContainerWorkdir(namespace),
			VoLabelKeyStageOrder:    fmt.Sprintf("%d", task.Extra.StageOrder),
		},
	}
	if volumeID != nil {
		vo.Labels[VoLabelKeyVolumeID] = *volumeID
	}
	return vo
}

func GenerateLocalVolume(namespace string, volumeID *string) apistructs.MetadataField {
	vo := apistructs.MetadataField{
		Name:  namespace,
		Value: ContainerContextDir,
		Type:  string(spec.StoreTypeDiceVolumeLocal),
		Labels: map[string]string{
			VoLabelKeyShareVolume: strconv.FormatBool(true),
		},
	}
	if volumeID != nil {
		vo.Labels[VoLabelKeyVolumeID] = *volumeID
	}
	return vo
}

func GenerateFakeVolume(namespace string, mountPath string, volumeID *string) apistructs.MetadataField {
	vo := apistructs.MetadataField{
		Name:  namespace,
		Value: mountPath,
		Type:  string(spec.StoreTypeDiceVolumeFake),
		Labels: map[string]string{
			VoLabelKeyContextPath:   mountPath,
			VoLabelKeyContainerPath: mountPath,
		},
	}
	if volumeID != nil {
		vo.Labels[VoLabelKeyVolumeID] = *volumeID
	}
	return vo
}
