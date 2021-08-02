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
	logrus.Infof("delete service group %s on namespace %s from etcd", sg.ID, ns)
	if err := s.js.Remove(context.Background(), mkServiceGroupKey(namespace, name), nil); err != nil {
		if force != "true" {
			return err
		}
	}
	logrus.Infof("delete service group %s on namespace %s from etcd successfully", sg.ID, ns)
	return nil
}
