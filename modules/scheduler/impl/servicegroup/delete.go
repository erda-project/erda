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
	"errors"

	"github.com/sirupsen/logrus"

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
				logrus.Errorf("not found runtime %s on namespace %s", name, namespace)
				return DeleteNotFound
			}
			logrus.Errorf("get from etcd err: %v when delete runtime %s on namespace %s", err, name, namespace)
			return err
		}
	}
	ns := sg.ProjectNamespace
	if ns == "" {
		ns = sg.Type
	}
	logrus.Infof("start to delete service group %s on namespace %s", sg.ID, ns)
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
	logrus.Infof("delete service group %s on namespace %s successfully", sg.ID, ns)
	return nil
}
