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

package instanceinfo

import (
	"github.com/erda-project/erda/apistructs"
	insinfo "github.com/erda-project/erda/modules/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type ComponentInfo interface {
	Get() (apistructs.ComponentInfoDataList, error)
}

type ComponentInfoImpl struct {
	db *insinfo.Client
}

func NewComponentInfoImpl() *ComponentInfoImpl {
	return &ComponentInfoImpl{
		db: insinfo.New(dbengine.MustOpen()),
	}
}

func (c *ComponentInfoImpl) Get() (apistructs.ComponentInfoDataList, error) {
	r := c.db.InstanceReader()
	instances, err := r.ByMetaLike("dice_component=").Do()
	if err != nil {
		return apistructs.ComponentInfoDataList{}, err
	}
	result := apistructs.ComponentInfoDataList{}
	for _, ins := range instances {
		insInfo := apistructs.ComponentInfoData{}
		name, ok := ins.Metadata("dice_component")
		if !ok {
			continue
		}

		insInfo.Cluster = ins.Cluster
		insInfo.ComponentName = name
		insInfo.Phase = string(ins.Phase)
		insInfo.Message = ins.Message
		insInfo.ContainerID = ins.ContainerID
		insInfo.ContainerIP = ins.ContainerIP
		insInfo.HostIP = ins.HostIP
		insInfo.ExitCode = ins.ExitCode
		insInfo.CpuOrigin = ins.CpuOrigin
		insInfo.MemOrigin = ins.MemOrigin
		insInfo.CpuRequest = ins.CpuRequest
		insInfo.MemRequest = ins.MemRequest
		insInfo.CpuLimit = ins.CpuLimit
		insInfo.MemLimit = ins.MemLimit
		insInfo.Image = ins.Image
		insInfo.StartedAt = ins.StartedAt
		insInfo.FinishedAt = ins.FinishedAt
		result = append(result, insInfo)
	}
	return result, nil
}
