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

package types

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/actionexecutor/logic"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	EnvRetainNamespace = "RETAIN_NAMESPACE"
)

type K8sBaseExecutor interface {
	Kind() Kind
	Name() Name

	Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error)
	Delete(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error)
	// CleanUp clean up all resources under this namespace
	CleanUp(ctx context.Context, namespace string) error
}

type K8sExecutor struct {
	K8sBaseExecutor
	errWrapper *logic.ErrorWrapper
}

func NewK8sExecutor(exe K8sBaseExecutor) *K8sExecutor {
	return &K8sExecutor{
		K8sBaseExecutor: exe,
		errWrapper:      logic.NewErrorWrapper(exe.Name().String()),
	}
}

func (k *K8sExecutor) Exist(ctx context.Context, task *spec.PipelineTask) (created bool, started bool, err error) {
	statusDesc, err := k.Status(ctx, task)
	if err != nil {
		created = false
		started = false
		if strutil.Contains(err.Error(), "failed to inspect job, err: not found") {
			err = nil
			return
		}
		return
	}
	return logic.JudgeExistedByStatus(statusDesc)
}

func (k *K8sExecutor) Create(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "create job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	created, _, err := k.Exist(ctx, task)
	if err != nil {
		return nil, err
	}
	if created {
		logrus.Warnf("%s: task already created, taskInfo: %s", k.Kind().String(), logic.PrintTaskInfo(task))
	}
	return nil, nil
}

func (k *K8sExecutor) Update(ctx context.Context, task *spec.PipelineTask) (interface{}, error) {
	return nil, errors.Errorf("%s not support update operation", k.Kind().String())
}

func (k *K8sExecutor) Cancel(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "cancel job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	// TODO move all makeJobID to framework
	// now move makeJobID to framework may change task uuid in database
	// Restore the task uuid after remove, because gc will make the job id, but cancel don't make the job id
	oldUUID := task.Extra.UUID
	task.Extra.UUID = task_uuid.MakeJobID(task)
	d, err := k.Delete(ctx, task)
	task.Extra.UUID = oldUUID
	return d, err
}

func (k *K8sExecutor) Remove(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error) {
	defer k.errWrapper.WrapTaskError(&err, "remove job", task)
	if err := logic.ValidateAction(task); err != nil {
		return nil, err
	}
	task.Extra.UUID = task_uuid.MakeJobID(task)
	return k.Delete(ctx, task)
}

func (k *K8sExecutor) BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (data interface{}, err error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	task := tasks[0]
	defer k.errWrapper.WrapTaskError(&err, "batch delete job", task)
	namespaces := sets.NewString()
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = k.Delete(ctx, task)
		if err != nil {
			return nil, err
		}
		if !isRetainNamespace(task) {
			namespaces.Insert(task.Extra.Namespace)
		}
	}
	logrus.Infof("start to clean up namespaces: %s", strings.Join(namespaces.List(), ","))
	for namespace := range namespaces {
		if err := k.CleanUp(ctx, namespace); err != nil {
			logrus.Errorf("failed to clean up namespace: %s, err: %v", namespace, err)
			return nil, err
		}
		logrus.Infof("successfully clean up namespace: %s", namespace)
	}
	return nil, nil
}

func isRetainNamespace(task *spec.PipelineTask) bool {
	retainNamespaceEnv := task.Extra.PublicEnvs[EnvRetainNamespace]
	if len(retainNamespaceEnv) == 0 {
		return false
	}
	retainNamespace, err := strconv.ParseBool(retainNamespaceEnv)
	if err != nil {
		logrus.Debugf("parse bool err %v when clean up the namespace %s", err, task.Extra.Namespace)
		return false
	}
	return retainNamespace
}
