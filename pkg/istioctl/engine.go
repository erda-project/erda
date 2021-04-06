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

package istioctl

import (
	"context"

	"github.com/sirupsen/logrus"
	"istio.io/client-go/pkg/clientset/versioned"

	"github.com/erda-project/erda/apistructs"
)

type ExecResult int

const (
	ExecSuccess ExecResult = iota
	ExecSkip
	ExecComplete
	ExecError
)

var EmptyEngine IstioEngine

type ServiceOp string

type IstioEngine interface {
	OnServiceOperator(ServiceOp, *apistructs.Service) error
}

const (
	ServiceCreate ServiceOp = "create"
	ServiceUpdate           = "update"
	ServiceDelete           = "delete"
)

type IstioExecutor interface {
	GetName() string
	SetIstioClient(versioned.Interface)
	OnServiceCreate(context.Context, *apistructs.Service) (ExecResult, error)
	OnServiceUpdate(context.Context, *apistructs.Service) (ExecResult, error)
	OnServiceDelete(context.Context, *apistructs.Service) (ExecResult, error)
}

type DefaultEngine struct {
	executors []IstioExecutor
	ctx       context.Context
}

func NewDefaultEngine(executors ...IstioExecutor) DefaultEngine {
	ctx := context.Background()
	return DefaultEngine{
		executors: executors,
		ctx:       ctx,
	}
}

// OnServiceOperator
func (engine DefaultEngine) OnServiceOperator(op ServiceOp, svc *apistructs.Service) error {
	for _, executor := range engine.executors {
		var result ExecResult
		var err error
		switch op {
		case ServiceCreate:
			if svc.MeshEnable != nil && *svc.MeshEnable {
				result, err = executor.OnServiceCreate(engine.ctx, svc)
			}
		case ServiceUpdate:
			if svc.MeshEnable != nil && *svc.MeshEnable {
				result, err = executor.OnServiceUpdate(engine.ctx, svc)
			} else {
				// 关闭了 service mesh
				result, err = executor.OnServiceDelete(engine.ctx, svc)
			}
		case ServiceDelete:
			// 总是清理，允许 Not Found
			result, err = executor.OnServiceDelete(engine.ctx, svc)
		}
		if err != nil {
			logrus.Errorf("op:%s, svc:%s, error happened: %+v", op, svc.Name, err)
		}
		switch result {
		case ExecComplete:
			return nil
		case ExecError:
			return err
		case ExecSkip:
			logrus.Errorf("istio executor:%s skiped", executor.GetName())
		}
	}
	return nil
}
