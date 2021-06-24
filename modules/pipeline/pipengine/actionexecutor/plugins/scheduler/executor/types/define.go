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

package types

import (
	"context"
	"fmt"

	"github.com/gogap/errors"

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
