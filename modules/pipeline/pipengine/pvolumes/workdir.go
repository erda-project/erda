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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// GetAvailableTaskContainerWorkdirs 查询当前存在的所有 Workdir
func GetAvailableTaskContainerWorkdirs(tasks []spec.PipelineTask, currentTask spec.PipelineTask) map[string]string {
	workdirs := make(map[string]string)

	taskNameMap := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		taskNameMap[task.Name] = struct{}{}
	}

	// 当前所有可用的 containerPaths
	containerPaths := GetAvailableTaskContainerPaths(tasks, currentTask)

	// 过滤出 taskName
	for name, containerPath := range containerPaths {
		if _, ok := taskNameMap[name]; !ok {
			continue
		}
		workdirs[name] = containerPath
	}

	return workdirs
}

func GetAvailableTaskContainerPaths(tasks []spec.PipelineTask, currentTask spec.PipelineTask) map[string]string {
	containerPaths := make(map[string]string)

	// 当前所有可用的 volumes
	for _, vo := range GetAvailableTaskOutStorages(tasks) {
		// 可用的 vo labels 里包含必要信息
		if vo.Labels == nil {
			continue
		}
		containerPath, ok := vo.Labels[VoLabelKeyContainerPath]
		if !ok {
			continue
		}
		containerPaths[vo.Name] = containerPath
	}

	// 可以通过 ref 引用自己
	for _, namespace := range currentTask.Extra.Action.Namespaces {
		containerPaths[namespace] = MakeTaskContainerWorkdir(namespace)
	}

	return containerPaths
}

func GetAvailableTaskOutStorages(tasks []spec.PipelineTask) []apistructs.MetadataField {
	var volumes []apistructs.MetadataField
	for _, task := range tasks {
		for _, out := range task.Context.OutStorages {
			volumes = append(volumes, out)
		}
	}
	return volumes
}
