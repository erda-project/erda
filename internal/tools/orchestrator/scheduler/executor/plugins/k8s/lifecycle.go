// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
)

const (
	DefaultProdTerminationGracePeriodSeconds = 45
)

var (
	DefaultProdLifecyclePreStopHandler = &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"sh", "-c", "sleep 10"},
		},
	}
)

func (k *Kubernetes) AddLifeCycle(service *apistructs.Service, podSpec *corev1.PodSpec) {
	if podSpec == nil || len(podSpec.Containers) == 0 {
		return
	}

	workspace, _ := util.GetDiceWorkspaceFromEnvs(service.Env)
	if workspace.Equal(apistructs.ProdWorkspace) {
		podSpec.TerminationGracePeriodSeconds = pointer.Int64(DefaultProdTerminationGracePeriodSeconds)
		setPreStopHandler(&podSpec.Containers[0])
	}
}

func setPreStopHandler(container *corev1.Container) {
	if container.Lifecycle == nil {
		container.Lifecycle = &corev1.Lifecycle{
			PreStop: DefaultProdLifecyclePreStopHandler,
		}
		return
	}
	container.Lifecycle.PreStop = DefaultProdLifecyclePreStopHandler
}
