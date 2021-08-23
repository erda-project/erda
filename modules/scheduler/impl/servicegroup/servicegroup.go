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

package servicegroup

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/impl/cluster/clusterutil"
	"github.com/erda-project/erda/modules/scheduler/impl/clusterinfo"
	"github.com/erda-project/erda/modules/scheduler/task"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	runtimeNameFormat                = `^[a-zA-Z0-9\-]+$`
	runtimeFormater   *regexp.Regexp = regexp.MustCompile(runtimeNameFormat)

	LastRestartTimeKey      = "lastRestartTime"
	LastConfigUpdateTimeKey = "lastConfigUpdateTime"
)

type Namespace string

type ServiceGroup interface {
	Create(sg apistructs.ServiceGroupCreateV2Request) (apistructs.ServiceGroup, error)
	Update(sg apistructs.ServiceGroupUpdateV2Request) (apistructs.ServiceGroup, error)
	Restart(namespace string, name string) error
	Cancel(namespace string, name string) error
	Delete(namespace string, name, force string) error
	Info(ctx context.Context, namespace string, name string) (apistructs.ServiceGroup, error)
	Precheck(sg apistructs.ServiceGroupPrecheckRequest) (apistructs.ServiceGroupPrecheckData, error)
	ConfigUpdate(sg apistructs.ServiceGroup) error
	KillPod(ctx context.Context, namespace string, name string, podname string) error
	Scale(sg *apistructs.ServiceGroup) (apistructs.ServiceGroup, error)
}

type ServiceGroupImpl struct {
	js          jsonstore.JsonStore
	sched       *task.Sched
	clusterinfo clusterinfo.ClusterInfo
}

func NewServiceGroupImpl(js jsonstore.JsonStore, sched *task.Sched, clusterinfo clusterinfo.ClusterInfo) ServiceGroup {
	return &ServiceGroupImpl{js, sched, clusterinfo}
}

func (s ServiceGroupImpl) handleKillPod(ctx context.Context, sg *apistructs.ServiceGroup, containerID string) (task.TaskResponse, error) {
	var (
		result task.TaskResponse
		err    error
	)
	if err = setServiceGroupExecutorByCluster(sg, s.clusterinfo); err != nil {
		return result, err
	}
	t, err := s.sched.Send(ctx, task.TaskRequest{
		ExecutorKind: getServiceExecutorKindByName(sg.Executor),
		ExecutorName: sg.Executor,
		Action:       task.TaskKillPod,
		ID:           sg.ID,
		Spec:         containerID,
	})
	if err != nil {
		return result, err
	}
	if result = t.Wait(ctx); result.Err() != nil {
		return result, result.Err()
	}
	return result, nil
}

func (s ServiceGroupImpl) handleServiceGroup(ctx context.Context, sg *apistructs.ServiceGroup, taskAction task.Action) (task.TaskResponse, error) {
	var (
		result task.TaskResponse
		err    error
	)
	if err = setServiceGroupExecutorByCluster(sg, s.clusterinfo); err != nil {
		return result, err
	}

	t, err := s.sched.Send(ctx, task.TaskRequest{
		ExecutorKind: getServiceExecutorKindByName(sg.Executor),
		ExecutorName: sg.Executor,
		Action:       taskAction,
		ID:           sg.ID,
		Spec:         *sg,
	})
	if err != nil {
		return result, err
	}
	if result = t.Wait(ctx); result.Err() != nil {
		return result, result.Err()
	}
	return result, nil

}

///////////////////////////////////////////////////////////////////////////////
//                   util funcs
//  TODO: Delete related code such as marathon in the following code
///////////////////////////////////////////////////////////////////////////////

func getServiceExecutorKindByName(name string) string {
	e, err := executor.GetManager().Get(executortypes.Name(name))
	if err != nil {
		return conf.DefaultRuntimeExecutor()
	}
	return string(e.Kind())
}
func mkServiceGroupKey(namespace, name string) string {
	return filepath.Join("/dice/service", namespace, name)
}
func validateServiceGroupName(name string) bool {
	return len(name) > 0 && runtimeFormater.MatchString(name)
}
func validateServiceGroupNamespace(namespace string) bool {
	return len(namespace) > 0 && runtimeFormater.MatchString(namespace)
}

// When the executor of the servicegroup is empty, set the executor according to the cluster value
func setServiceGroupExecutorByCluster(sg *apistructs.ServiceGroup, clusterinfo clusterinfo.ClusterInfo) error {
	if len(sg.Executor) > 0 {
		return nil
	}
	if len(sg.ClusterName) == 0 {
		return errors.Errorf("servicegroup(%s/%s) neither executor nor cluster is set", sg.Type, sg.ID)
	}

	// Determine whether it is edas deployment addon scenario
	isEdasStatefull := isEdasStatefull(sg, clusterinfo)

	if isEdasStatefull {
		sg.Executor = clusterutil.GenerateExecutorByCluster(sg.ClusterName, clusterutil.EdasKindInK8s)
	} else {
		sg.Executor = clusterutil.GenerateExecutorByCluster(sg.ClusterName, clusterutil.ServiceKindMarathon)
	}
	logrus.Infof("generate executor(%s) for servicegroup(%s) in cluster(%s)", sg.Executor, sg.ID, sg.ClusterName)
	return nil
}

func isEdasStatefull(sg *apistructs.ServiceGroup, clusterinfo clusterinfo.ClusterInfo) bool {
	info, err := clusterinfo.Info(sg.ClusterName)
	if err != nil {
		logrus.Errorf("failed to get cluster info, clusterName: %s, (%v)", sg.ClusterName, err)
		return false
	}

	//Set the executor of edasink8s when the cluster type is edas and it is an addon type service
	if info.Get(apistructs.DICE_CLUSTER_TYPE) == "edas" {
		_, ok := sg.Labels["USE_OPERATOR"]
		if ok {
			return true
		}
		if sg.Labels["SERVICE_TYPE"] == "ADDONS" {
			return true
		}
	}
	return false
}

func convertServiceGroup(req apistructs.ServiceGroupCreateV2Request, clusterinfo clusterinfo.ClusterInfo) (apistructs.ServiceGroup, error) {
	sg := apistructs.ServiceGroup{}
	sg.ClusterName = req.ClusterName
	sg.Force = true
	if !validateServiceGroupName(req.ID) || !validateServiceGroupNamespace(req.Type) {
		return apistructs.ServiceGroup{}, errors.New("invalid Name or Namespace")
	}

	// build match tags and exclude tags
	sg.Labels = appendServiceTags(sg.Labels, sg.Executor)

	sg.ID = req.ID
	sg.Type = req.Type
	sg.ProjectNamespace = req.ProjectNamespace
	sg.Labels = req.GroupLabels
	sg.ServiceDiscoveryMode = req.ServiceDiscoveryMode

	yml := req.DiceYml
	// expand global envs to service's envs
	diceyml.ExpandGlobalEnv(&yml)
	reqVolumesInfo := req.Volumes

	// override sg.Labels
	if sg.Labels == nil {
		sg.Labels = map[string]string{}
	}
	for k, v := range yml.Meta {
		sg.Labels[k] = v
	}
	// override sg.ServiceDiscoveryMode
	if mode, ok := yml.Meta["SERVICE_DISCOVERY_MODE"]; ok {
		sg.ServiceDiscoveryMode = strutil.ToUpper(mode)
	}

	sgServices := []apistructs.Service{}
	for name, service := range yml.Services {
		binds := []apistructs.ServiceBind{}
		ymlbinds, err := diceyml.ParseBinds(service.Binds)
		if err != nil {
			return apistructs.ServiceGroup{}, err
		}
		for _, bind := range ymlbinds {
			ro := true
			if bind.Type == "rw" {
				ro = false
			}
			binds = append(binds, apistructs.ServiceBind{
				Bind: apistructs.Bind{
					ContainerPath: bind.ContainerPath,
					HostPath:      bind.HostPath,
					ReadOnly:      ro,
				},
			})
		}

		volumes := []apistructs.Volume{}
		volumeInfo, ok := reqVolumesInfo[name]
		if ok {
			var tp apistructs.VolumeType
			var err error
			if volumeInfo.Type == "" {
				tp = apistructs.LocalVolume
			} else {
				tp, err = apistructs.VolumeTypeFromString(volumeInfo.Type)
				if err != nil {
					return apistructs.ServiceGroup{}, fmt.Errorf("bad volume type: %v", volumeInfo.Type)

				}
			}
			volumes = append(volumes, apistructs.Volume{
				ID:            volumeInfo.ID,
				VolumeType:    tp,
				Size:          10,
				ContainerPath: volumeInfo.ContainerPath,
			})
		}

		sgService := apistructs.Service{
			Name:          name,
			Image:         service.Image,
			ImageUsername: service.ImageUsername,
			ImagePassword: service.ImagePassword,
			Cmd:           service.Cmd,
			Ports:         service.Ports,
			Scale:         service.Deployments.Replicas,
			Resources: apistructs.Resources{
				Cpu:  service.Resources.CPU,
				Mem:  float64(service.Resources.Mem),
				Disk: float64(service.Resources.Disk),
			},
			Depends:          service.DependsOn,
			Env:              service.Envs,
			Labels:           service.Labels,
			Selectors:        service.Deployments.Selectors,
			WorkLoad:         service.Deployments.Workload,
			DeploymentLabels: service.Deployments.Labels,
			Binds:            binds,
			Volumes:          volumes,
			Hosts:            service.Hosts,
			NewHealthCheck:   convertHealthcheck(service.HealthCheck),
			SideCars:         service.SideCars,
			InitContainer:    service.Init,
			MeshEnable:       service.MeshEnable,
			TrafficSecurity:  service.TrafficSecurity,
			K8SSnippet:       service.K8SSnippet,
		}
		sgServices = append(sgServices, sgService)
	}
	sg.Services = sgServices
	if err := setServiceGroupExecutorByCluster(&sg, clusterinfo); err != nil {
		return apistructs.ServiceGroup{}, err
	}
	return sg, nil
}
func convertHealthcheck(hc diceyml.HealthCheck) *apistructs.NewHealthCheck {
	nhc := apistructs.NewHealthCheck{}
	if hc.HTTP != nil && hc.HTTP.Port != 0 && hc.HTTP.Path != "" {
		nhc.HttpHealthCheck = &apistructs.HttpHealthCheck{
			Port:     hc.HTTP.Port,
			Path:     hc.HTTP.Path,
			Duration: hc.HTTP.Duration,
		}
	}
	if hc.Exec != nil && hc.Exec.Cmd != "" {
		nhc.ExecHealthCheck = &apistructs.ExecHealthCheck{
			Cmd:      hc.Exec.Cmd,
			Duration: hc.Exec.Duration,
		}
	}
	return &nhc
}

func appendServiceTags(labels map[string]string, executor string) map[string]string {
	matchTags := make([]string, 0)
	if labels["SERVICE_TYPE"] == "STATELESS" {
		matchTags = append(matchTags, apistructs.TagServiceStateless)
	} else if executor == "MARATHONFORTERMINUSY" && labels["SERVICE_TYPE"] == "ADDONS" {
		matchTags = append(matchTags, apistructs.TagServiceStateful)
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[apistructs.LabelMatchTags] = strings.Join(matchTags, ",")
	labels[apistructs.LabelExcludeTags] = apistructs.TagLocked + "," + apistructs.TagPlatform
	return labels
}
