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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Info(ctx context.Context, namespace string, name string) (apistructs.ServiceGroup, error) {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return sg, err
	}

	result, err := s.handleServiceGroup(ctx, &sg, task.TaskInspect)
	if err != nil {
		return sg, err
	}
	if result.Extra == nil {
		err = errors.Errorf("Cannot get servicegroup(%v/%v) info from TaskInspect", sg.Type, sg.ID)
		logrus.Error(err.Error())
		return sg, err
	}

	newsg := result.Extra.(*apistructs.ServiceGroup)
	return *newsg, nil
}
