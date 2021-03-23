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
