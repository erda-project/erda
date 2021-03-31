package servicegroup

import (
	"context"
	"errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/jsonstore"
)

var DeleteNotFound error = errors.New("not found")

func (s ServiceGroupImpl) Delete(namespace string, name, force string) error {
	sg := apistructs.ServiceGroup{}
	// force offline, first time not set status offline, delete etcd data; after set status, get and delete again, not return error
	if err := s.js.Get(context.Background(), mkServiceGroupKey(namespace, name), &sg); err != nil {
		if force != "true" {
			if err == jsonstore.NotFoundErr {
				return DeleteNotFound
			}
			return err
		}
	}

	if _, err := s.handleServiceGroup(context.Background(), &sg, task.TaskDestroy); err != nil {
		if force != "true" {
			return err
		}
	}
	if err := s.js.Remove(context.Background(), mkServiceGroupKey(namespace, name), nil); err != nil {
		if force != "true" {
			return err
		}
	}
	return nil
}
