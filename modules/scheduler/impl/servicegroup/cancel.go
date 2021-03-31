package servicegroup

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
)

// TODO: 对比service_endpoints.go, 返回的内容是否应该改一下？
func (s ServiceGroupImpl) Cancel(namespace string, name string) error {
	sg := apistructs.ServiceGroup{}
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		return err
	}

	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskCancel); err != nil {
		return err
	}
	return nil
}
