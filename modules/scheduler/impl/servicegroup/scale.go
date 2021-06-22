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

package servicegroup

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s *ServiceGroupImpl) Scale(sg *apistructs.ServiceGroup) (apistructs.ServiceGroup, error) {
	logrus.Infof("start to scale service %v", sg.Services)
	oldSg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, fmt.Errorf("Cannot get servicegroup(%s/%s) from etcd, err: %v", sg.Type, sg.ID, err)
	}

	if len(sg.Services) != 1 {
		return apistructs.ServiceGroup{}, fmt.Errorf("services count more than 1")
	}

	// get sg info from etcd storage, and set the project namespace to the scale sg
	// when the project namespace is not empty
	if oldSg.ProjectNamespace != "" {
		sg.ProjectNamespace = oldSg.ProjectNamespace
	}

	newService := sg.Services[0]
	_, err := s.handleServiceGroup(context.Background(), sg, task.TaskScale)
	if err != nil {
		errMsg := fmt.Sprintf("scale service %v err: %v", sg.Services, err)
		logrus.Error(errMsg)
		return *sg, fmt.Errorf(errMsg)
	}
	for index, svc := range oldSg.Services {
		if svc.Name == newService.Name {
			if svc.Scale != newService.Scale {
				svc.Scale = newService.Scale
			}
			if svc.Resources.Cpu != newService.Resources.Cpu || svc.Resources.Mem != newService.Resources.Mem {
				svc.Resources = newService.Resources
			}
			oldSg.Services[index] = svc
			break
		}
	}
	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &oldSg); err != nil {
		return apistructs.ServiceGroup{}, err
	}
	return *sg, nil
}
