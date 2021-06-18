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

func (s *ServiceGroupImpl) Scale(sg *apistructs.ServiceGroup) error {
	logrus.Info("start to scale service %v", sg.Services)
	_, err := s.handleServiceGroup(context.Background(), sg, task.TaskScale)
	if err != nil {
		errMsg := fmt.Sprintf("scale service %v err: %v", sg.Services, err)
		logrus.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}
