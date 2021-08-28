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

package containers

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func GenContainers(task *spec.PipelineTask) ([]apistructs.TaskContainer, error) {
	if value, ok := task.Extra.Action.Params["bigDataConf"]; ok {
		spec := apistructs.BigdataSpec{}
		if err := json.Unmarshal([]byte(value.(string)), &spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task bigDataConf")
		}
		if spec.FlinkConf != nil {
			return GenFlinkContainers(task), nil
		}
		//if spec.SparkConf != nil {
		//	executorName = "k8sspark"
		//}
	}
	return GenTaskContainer(task), nil
}

func GenTaskContainer(task *spec.PipelineTask) []apistructs.TaskContainer {
	return []apistructs.TaskContainer{apistructs.TaskContainer{
		TaskName:    task.Name,
		ContainerID: task.Extra.UUID,
	}}
}

func MakeFlinkJobName(name string) string {
	return fmt.Sprintf("%s-job", name)
}

func MakeFlinkJobID(uuid string) string {
	return fmt.Sprintf("%s-job", uuid)
}

func MakeFlinkJobManagerName(name string) string {
	return fmt.Sprintf("%s-job-manager", name)
}

func MakeFlinkJobManagerID(uuid string) string {
	return fmt.Sprintf("%s-job-manager", uuid)
}

func MakeFlinkTaskManagerName(name string) string {
	return fmt.Sprintf("%s-task-manager", name)
}

func MakeFLinkTaskManagerID(uuid string) string {
	return fmt.Sprintf("%s-task-manager", uuid)
}

func GenFlinkContainers(task *spec.PipelineTask) []apistructs.TaskContainer {
	containers := make([]apistructs.TaskContainer, 0)
	containers = append(containers, apistructs.TaskContainer{
		TaskName:    MakeFlinkJobName(task.Name),
		ContainerID: MakeFlinkJobID(task.Extra.UUID),
	})
	containers = append(containers, apistructs.TaskContainer{
		TaskName:    MakeFlinkJobManagerName(task.Name),
		ContainerID: MakeFlinkJobManagerID(task.Extra.UUID),
	})
	containers = append(containers, apistructs.TaskContainer{
		TaskName:    MakeFlinkTaskManagerName(task.Name),
		ContainerID: MakeFLinkTaskManagerID(task.Extra.UUID),
	})
	return containers
}
