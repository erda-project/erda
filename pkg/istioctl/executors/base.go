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

package executors

import (
	"context"

	"istio.io/client-go/pkg/clientset/versioned"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/istioctl"
)

type BaseExecutor struct {
	client versioned.Interface
}

// SetIstioClient
func (exe *BaseExecutor) SetIstioClient(client versioned.Interface) {
	exe.client = client
}

// OnServiceCreate
func (exe BaseExecutor) OnServiceCreate(context.Context, *apistructs.Service) (istioctl.ExecResult, error) {
	return istioctl.ExecSuccess, nil
}

// OnServiceUpdate
func (exe BaseExecutor) OnServiceUpdate(context.Context, *apistructs.Service) (istioctl.ExecResult, error) {
	return istioctl.ExecSuccess, nil
}

// OnServiceDelete
func (exe BaseExecutor) OnServiceDelete(context.Context, *apistructs.Service) (istioctl.ExecResult, error) {
	return istioctl.ExecSuccess, nil
}
