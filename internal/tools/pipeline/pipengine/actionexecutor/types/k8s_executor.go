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
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/pipengine/actionexecutor/logic"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/k8sclient"
	"github.com/erda-project/erda/pkg/strutil"
)

type K8sBaseExecutor interface {
	Kind() Kind
	Name() Name

	Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error)
	Delete(ctx context.Context, task *spec.PipelineTask) (data interface{}, err error)
}

type K8sExecutor struct {
	sync.Mutex

	K8sBaseExecutor
	Client *k8sclient.K8sClient

	errWrapper *logic.ErrorWrapper
	StopCh     chan struct{}
	handlers   map[string][]reflect.Value
}

func NewK8sExecutor(clusterName string, exe K8sBaseExecutor) (*K8sExecutor, error) {
	// we could operate normal resources (job, pod, deploy,pvc,pv,crd and so on) by default config permissions(injected by kubernetes, /var/run/secrets/kubernetes.io/serviceaccount)
	// so WithPreferredToUseInClusterConfig it's enough for pipeline and orchestrator
	client, err := k8sclient.New(clusterName, k8sclient.WithTimeout(time.Duration(conf.K8SExecutorMaxInitializationSec())*time.Second), k8sclient.WithPreferredToUseInClusterConfig())
	if err != nil {
		return nil, err
	}
	return &K8sExecutor{
		StopCh:          make(chan struct{}),
		Client:          client,
		K8sBaseExecutor: exe,
		errWrapper:      logic.NewErrorWrapper(exe.Name().String()),
		handlers:        map[string][]reflect.Value{},
	}, nil
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
	for _, task := range tasks {
		if len(task.Extra.UUID) <= 0 {
			continue
		}
		_, err = k.Delete(ctx, task)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (k *K8sExecutor) SubscribeEvent(ctx context.Context, task *spec.PipelineTask, f interface{}) error {
	k.Lock()
	defer k.Unlock()

	v := reflect.ValueOf(f)
	if v.Type().Kind() != reflect.Func {
		return fmt.Errorf("handler: %v is not a function", v.Type().Kind().String())
	}

	handlers, ok := k.handlers[MakeJobName(task.Extra.Namespace, task.Extra.UUID)]
	if !ok {
		handlers = []reflect.Value{}
	}
	handlers = append(handlers, v)
	k.handlers[MakeJobName(task.Extra.Namespace, task.Extra.UUID)] = handlers
	return nil
}

func (k *K8sExecutor) PublishEvent(identity string, args ...interface{}) {
	handlers, ok := k.handlers[identity]
	if !ok {
		return
	}

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}
	for i := range handlers {
		go func(idx int) {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("%s: handler: %s, error: %s", k.Kind().String(), handlers[idx].Type().String(), err)
				}
			}()
			handlers[idx].Call(params)
		}(i)
	}
}

func MakeJobName(namespace string, taskUUID string) string {
	return strutil.Concat(namespace, ".", taskUUID)
}
