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

package demo

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const (
	Kind = "DEMO"
)

func init() {
	types.Register("DEMO", func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		return &Demo{
			name: name,
		}, nil
	})
}

type Demo struct {
	name types.Name
}

func (d *Demo) Kind() types.Kind {
	return Kind
}

func (d *Demo) Name() types.Name {
	return d.name
}

func (d *Demo) Exist(ctx context.Context, action *spec.PipelineTask) (bool, bool, error) {
	return true, true, nil
}

func (d *Demo) Create(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo create ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Start(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo start ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Update(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo update ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Status(ctx context.Context, action *spec.PipelineTask) (apistructs.PipelineStatusDesc, error) {
	logrus.Info("demo status ...")
	time.Sleep(time.Second * 5)
	return apistructs.PipelineStatusDesc{Status: apistructs.PipelineStatusQueue, Desc: "lack of machine resource, waiting ..."}, nil
}

func (d *Demo) Inspect(ctx context.Context, action *spec.PipelineTask) (apistructs.TaskInspect, error) {
	logrus.Info("demo inspect ...")
	time.Sleep(time.Second * 5)
	return apistructs.TaskInspect{}, nil
}

func (d *Demo) Cancel(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo cancel ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) Remove(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo remove ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}

func (d *Demo) BatchDelete(ctx context.Context, actions []*spec.PipelineTask) (interface{}, error) {
	logrus.Info("demo batch delete ...")
	time.Sleep(time.Second * 5)
	return nil, nil
}
