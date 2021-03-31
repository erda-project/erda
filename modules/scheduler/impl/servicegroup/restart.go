package servicegroup

import (
	"context"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

func (s ServiceGroupImpl) Restart(namespace string, name string) error {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return err
	}
	sg.Extra[LastRestartTimeKey] = time.Now().String()
	sg.LastModifiedTime = time.Now().Unix()

	if err := s.js.Put(context.Background(), mkServiceGroupKey(sg.Type, sg.ID), &sg); err != nil {
		return err
	}

	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)

	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskUpdate); err != nil {
		return err
	}
	return nil
}
