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

func (m *Memory) Inspect(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Cancel(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Remove(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) Destroy(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}

func (m *Memory) DeleteNamespace(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	panic("implement me")
}
