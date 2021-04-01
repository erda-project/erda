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
