package executor

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"

	"github.com/sirupsen/logrus"
)

type Empty struct {
	ExecutorName executortypes.Name `json:"name,omitempty"`
	ExecutorKind executortypes.Kind `json:"kind,omitempty"`
}

func (e *Empty) Kind() executortypes.Kind {
	logrus.Infof("Kind on empty")
	return e.ExecutorKind
}

func (e *Empty) Name() executortypes.Name {
	logrus.Infof("Name on empty")
	return e.ExecutorName
}

func (e *Empty) Create(ctx context.Context, specObj interface{}) (interface{}, error) {
	logrus.Infof("Create: %v", specObj)
	return nil, nil
}

func (e *Empty) Destroy(ctx context.Context, spec interface{}) error {
	logrus.Infof("Destroy on empty")
	return nil
}

func (e *Empty) Status(ctx context.Context, specObj interface{}) (apistructs.StatusDesc, error) {
	logrus.Infof("Status on empty")
	return apistructs.StatusDesc{}, nil
}

func (e *Empty) Remove(ctx context.Context, spec interface{}) error {
	logrus.Infof("Remove on empty")
	return nil
}

func (e *Empty) Update(ctx context.Context, specObj interface{}) (interface{}, error) {
	logrus.Infof("Update: %v", specObj)
	return nil, nil
}

func (e *Empty) Inspect(ctx context.Context, spec interface{}) (interface{}, error) {
	logrus.Infof("Inspect on empty")
	return nil, nil
}

func (e *Empty) Cancel(ctx context.Context, spec interface{}) (interface{}, error) {
	logrus.Infof("Cancel on empty")
	return nil, nil
}
