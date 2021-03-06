// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
