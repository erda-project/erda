package servicegroup

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
