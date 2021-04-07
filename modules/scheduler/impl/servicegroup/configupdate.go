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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) ConfigUpdate(sg apistructs.ServiceGroup) error {

	sg.LastModifiedTime = time.Now().Unix()

	logrus.Debugf("config update sg: %+v", sg)
	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &sg); err != nil {
		return err
	}

	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)

	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskUpdate); err != nil {
		return err
	}
	return nil
}
