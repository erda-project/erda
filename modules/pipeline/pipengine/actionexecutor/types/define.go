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
	"regexp"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type ActionExecutor interface {
	Kind() Kind
	Name() Name

	// Exist 返回 created, started, error
	Exist(ctx context.Context, action *spec.PipelineTask) (created bool, started bool, err error)
	// Create 保证幂等
	Create(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	// Start 保证幂等
	Start(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	Update(ctx context.Context, action *spec.PipelineTask) (interface{}, error)

	// Status 只做简单重试
	Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error)
	Inspect(ctx context.Context, action *spec.PipelineTask) (apistructs.TaskInspect, error)

	Cancel(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	Remove(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	BatchDelete(ctx context.Context, actions []*spec.PipelineTask) (interface{}, error)
}

const kindNameFormat = `^[A-Z0-9]+$`

var formatter = regexp.MustCompile(kindNameFormat)

// Name represents an executor's name.
type Name string

func (s Name) String() string {
	return string(s)
}

func (s Name) Validate() bool {
	return formatter.MatchString(string(s))
}

// Kind represents an executor's type.
type Kind string

func (s Kind) String() string {
	return string(s)
}

func (s Kind) Validate() bool {
	return formatter.MatchString(string(s))
}

// Create be used to create an action executor instance.
type CreateFn func(name Name, options map[string]string) (ActionExecutor, error)

var Factory = map[Kind]CreateFn{}

// Register add an executor's create function.
func Register(kind Kind, create CreateFn) error {
	if !kind.Validate() {
		return errors.Errorf("invalid kind: %s", kind)
	}
	if _, ok := Factory[kind]; ok {
		return errors.Errorf("duplicate to register action executor: %s", kind)
	}
	Factory[kind] = create
	return nil
}

// MustRegister panic if register failed.
func MustRegister(kind Kind, create CreateFn) {
	err := Register(kind, create)
	if err != nil {
		panic(fmt.Errorf("failed to register action executor, kind: %s, err: %v", kind, err))
	}
}
