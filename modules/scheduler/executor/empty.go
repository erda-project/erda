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

package executor

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
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
