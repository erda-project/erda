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

package memory

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const (
	Kind = "MEMORY"
)

func init() {
	types.Register(Kind, func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		return &Memory{
			name:    name,
			options: options,
		}, nil
	})
}

type Memory struct {
	name    types.Name
	options map[string]string
}

func (m *Memory) Kind() types.Kind {
	return Kind
}

func (m *Memory) Name() types.Name {
	return m.name
}

func (m *Memory) Exist(ctx context.Context, action *spec.PipelineTask) (bool, bool, error) {
	panic("implement me")
}

func (m *Memory) Create(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Start(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Update(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error) {
	panic("implement me")
}

func (m *Memory) Inspect(ctx context.Context, action *spec.PipelineTask) (apistructs.TaskInspect, error) {
	panic("implement me")
}

func (m *Memory) Cancel(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Remove(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) BatchDelete(ctx context.Context, actions []*spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}
