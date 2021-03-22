package types

import (
	"context"
	"regexp"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type ActionExecutor interface {
	Kind() Kind
	Name() Name

	// Exist 返回 created, started, error
	Exist(ctx context.Context, action *spec.PipelineTask) (bool, bool, error)
	// Create 保证幂等
	Create(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	// Start 保证幂等
	Start(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	Update(ctx context.Context, action *spec.PipelineTask) (interface{}, error)

	// Status 只做简单重试
	Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error)
	Inspect(ctx context.Context, action *spec.PipelineTask) (interface{}, error)

	Cancel(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	Remove(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	Destroy(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
	DeleteNamespace(ctx context.Context, action *spec.PipelineTask) (interface{}, error)
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
