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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// Kind return the task executor type
type Kind string

func (k Kind) String() string {
	return string(k)
}

// Name return the task executor name
type Name string

func (n Name) String() string {
	return string(n)
}

type TaskExecutor interface {
	Kind() Kind
	Name() Name

	Status(ctx context.Context, task *spec.PipelineTask) (apistructs.StatusDesc, error)
	Create(ctx context.Context, task *spec.PipelineTask) (interface{}, error)
	Remove(ctx context.Context, task *spec.PipelineTask) (interface{}, error)
	BatchDelete(ctx context.Context, tasks []*spec.PipelineTask) (interface{}, error)
	Inspect(ctx context.Context, task *spec.PipelineTask) (apistructs.TaskInspect, error)
}

type CreateFn func(name Name, clusterName string, cluster apistructs.ClusterInfo) (TaskExecutor, error)

var Factory = map[Kind]CreateFn{}

func Register(kind Kind, create CreateFn) error {
	if _, ok := Factory[kind]; ok {
		return errors.Errorf("duplicate to register task executor: %s", kind)
	}
	Factory[kind] = create
	return nil
}

func MustRegister(kind Kind, create CreateFn) {
	err := Register(kind, create)
	if err != nil {
		panic(fmt.Errorf("failed to register action executor, kind: %s, err: %v", kind, err))
	}
}
