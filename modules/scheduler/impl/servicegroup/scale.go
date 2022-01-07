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

package servicegroup

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s *ServiceGroupImpl) Scale(sg *apistructs.ServiceGroup) (apistructs.ServiceGroup, error) {
	oldSg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, fmt.Errorf("Cannot get servicegroup(%s/%s) from etcd, err: %v", sg.Type, sg.ID, err)
	}

	/*
		if len(sg.Services) != 1 {
			return apistructs.ServiceGroup{}, fmt.Errorf("services count more than 1")
		}
	*/
	// get sg info from etcd storage, and set the project namespace to the scale sg
	// when the project namespace is not empty
	if oldSg.ProjectNamespace != "" {
		sg.ProjectNamespace = oldSg.ProjectNamespace
	}

	//newService := sg.Services[0]
	oldServiceReplicas := make(map[string]int)
	for index, svc := range oldSg.Services {
		for newIndex, newSvc := range sg.Services {
			if svc.Name == newSvc.Name {
				if svc.Scale != newSvc.Scale {
					if newSvc.Scale == 0 {
						oldServiceReplicas[svc.Name] = svc.Scale
					}
					svc.Scale = newSvc.Scale
				}
				if svc.Resources.Cpu != newSvc.Resources.Cpu || svc.Resources.Mem != newSvc.Resources.Mem {
					svc.Resources = newSvc.Resources
				}
				oldSg.Services[index] = svc
				sg.Services[newIndex] = oldSg.Services[index]
				break
			}
		}
		/*
			if svc.Name == newService.Name {
				if svc.Scale != newService.Scale {
					svc.Scale = newService.Scale
				}
				if svc.Resources.Cpu != newService.Resources.Cpu || svc.Resources.Mem != newService.Resources.Mem {
					svc.Resources = newService.Resources
				}
				oldSg.Services[index] = svc
				sg.Services[0] = oldSg.Services[index]
				break
			}
		*/
	}
	_, err := s.handleServiceGroup(context.Background(), sg, task.TaskScale)
	if err != nil {
		logrus.Errorf("scale service %s err: %v", sg.ID, err)
		return *sg, err
	}

	// 如果目前操作是停止(副本数为0)，为了后续恢复，需要保留停止操作前的副本数
	for index, oldSvc := range oldSg.Services {
		if replicas, ok := oldServiceReplicas[oldSvc.Name]; ok {
			oldSg.Services[index].Scale = replicas
		}
	}

	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, err
	}
	return *sg, nil
}
