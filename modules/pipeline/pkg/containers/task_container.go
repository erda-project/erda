// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	return []apistructs.TaskContainer{{
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
