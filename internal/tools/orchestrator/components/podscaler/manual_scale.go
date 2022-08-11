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

package podscaler

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// processRuntimeScaleRecord 处理单个 runtime 对应的 scale 操作
func (s *podscalerService) processRuntimeScaleRecord(rsc pb.RuntimeScaleRecord, action string) (*pb.PreDiceDTO, error) {
	uniqueId := spec.RuntimeUniqueId{
		ApplicationId: rsc.ApplicationId,
		Workspace:     rsc.Workspace,
		Name:          rsc.Name,
	}

	logrus.Infof("[processRuntimeScaleRecord] process runtime scale for runtime %#v", uniqueId)
	hpaRules, err := s.db.GetRuntimeHPARulesByServices(uniqueId, nil)
	if err != nil {
		return nil, errors.Errorf("[processRuntimeScaleRecord] get hpa rule failed for runtime uniqueID %+v, error: %v", uniqueId, err)
	}
	appliedScaledObjects := make(map[string]string)
	for _, rule := range hpaRules {
		// only applied rules need to delete
		if rule.IsApplied == pstypes.RuntimePARuleApplied {
			appliedScaledObjects[rule.ServiceName] = rule.Rules
		}
	}

	for svcName := range rsc.Payload.Services {
		if _, ok := appliedScaledObjects[svcName]; ok {
			return nil, errors.Errorf("[processRuntimeScaleRecord] hpa rules has applied for RuntimeUniqueId %#v, can not do this scale action, please delete or canel the applied rules first", uniqueId)
		}
	}

	pre, err := s.db.GetPreDeployment(uniqueId)
	if err != nil {
		logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, find PreDeployment record in table ps_v2_pre_builds failed for runtime %#v, err: %v", uniqueId, err)
		return nil, err
	}

	oldOverlay, err := getPreDeploymentDiceOverlay(pre)
	if err != nil {
		return nil, err
	}

	// Global Envs
	if rsc.Payload.Envs != nil {
		oldOverlay.Envs = rsc.Payload.Envs
	}
	var (
		needUpdateServices     []string
		oldOverlayDataForAudit = &pb.PreDiceDTO{
			Services: make(map[string]*pb.RuntimeInspectServiceDTO),
		}
	)
	//oldServiceReplicas := make(map[string]int)
	for k, v := range rsc.Payload.Services {
		oldService, exists := oldOverlay.Services[k]
		if !exists || oldService == nil {
			return nil, errors.Errorf("service %s is not found in ps_v2_pre_builds in dice filed or invalid service object", k)
		}
		// Local Envs
		if v.Envs != nil {
			oldService.Envs = v.Envs
		}

		oldOverlayDataForAudit.Services[k] = genOverlayDataForAudit(oldService)
		needUpdateServices = append(needUpdateServices, k)
		// Replicas
		oldService.Deployments.Replicas = int(v.Deployments.Replicas)
		oldOverlay.Services[k].Deployments.Replicas = int(v.Deployments.Replicas)
		// Resources
		oldService.Resources.CPU = v.Resources.Cpu
		oldOverlay.Services[k].Resources.CPU = v.Resources.Cpu
		oldService.Resources.Mem = int(v.Resources.Mem)
		oldOverlay.Services[k].Resources.Mem = int(v.Resources.Mem)
	}

	// really update scale
	if len(needUpdateServices) != 0 {
		runtime, err := s.db.GetRuntimeByUniqueID(uniqueId)
		if err != nil {
			logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, find runtime record in ps_v2_project_runtimes for runtime %#v failed, err: %v", uniqueId, err)
			return nil, err
		}
		if runtime == nil {
			logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, runtime %s is not existed for runtime %#v", uniqueId.Name, uniqueId)
			return nil, errors.Errorf("[processRuntimeScaleRecord] runtime %s is not existed", uniqueId.Name)
		}

		namespace, name := runtime.ScheduleName.Args()
		sg := &apistructs.ServiceGroup{
			ClusterName: runtime.ClusterName,
			Dice: apistructs.Dice{
				ID:       name,
				Type:     namespace,
				Services: make([]apistructs.Service, 0),
			},
		}
		for _, svcName := range needUpdateServices {
			sg.Services = append(sg.Services, apistructs.Service{
				Name:  svcName,
				Scale: oldOverlay.Services[svcName].Deployments.Replicas,
				Resources: apistructs.Resources{
					Cpu:  oldOverlay.Services[svcName].Resources.CPU,
					Mem:  float64(oldOverlay.Services[svcName].Resources.Mem),
					Disk: float64(oldOverlay.Services[svcName].Resources.Disk),
				},
			})
		}

		sgb, _ := json.Marshal(&sg)
		logrus.Debugf("[processRuntimeScaleRecord] scale service group body is %s", string(sgb))

		if _, err := s.serviceGroupImpl.Scale(sg); err != nil {
			logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, ScaleServiceGroup failed for runtime %s for runtime %#v failed, err: %v", uniqueId.Name, uniqueId, err)
			return nil, err
		}
	}

	// save changes
	o_, err := json.Marshal(oldOverlay)
	if err != nil {
		logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, ScaleServiceGroup for runtime %s for runtime %#v successfully, but json.Marshal diceyml.Object oldOverlay %#v failed, err: %v", uniqueId.Name, uniqueId, oldOverlay, err)
		return nil, err
	}
	pre.DiceOverlay = string(o_)
	logrus.Infof("pre is %+v", pre)
	if err := s.db.UpdatePreDeployment(pre); err != nil {
		logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, ScaleServiceGroup for runtime %s for runtime %#v successfully, but update table ps_v2_pre_builds failed, err: %v", uniqueId.Name, uniqueId, err)
		return nil, err
	}

	logrus.Infof("[processRuntimeScaleRecord] process runtime scale for for services %v successfully", needUpdateServices)
	return oldOverlayDataForAudit, nil
}

func getPreDeploymentDiceOverlay(pre *dbclient.PreDeployment) (diceyml.Object, error) {
	var oldOverlay diceyml.Object
	var diceObj diceyml.Object
	// 没有 pre.DiceOverlay 信息表示部署之后还未进行过 scale 操作，因此如果当前这次 scale 操作是 scaleDown, 则按 pre.Dice 中的副本数设置 恢复时的副本数
	if err := json.Unmarshal([]byte(pre.Dice), &diceObj); err != nil {
		logrus.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
		return oldOverlay, errors.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
	}
	oldOverlay.Services = make(map[string]*diceyml.Service)
	for k, v := range diceObj.Services {
		oldOverlay.Services[k] = &diceyml.Service{
			Deployments: diceyml.Deployments{
				Replicas: v.Deployments.Replicas,
			},
			Resources: diceyml.Resources{
				CPU: v.Resources.CPU,
				Mem: v.Resources.Mem,
			},
		}
	}

	var currentDiceObj diceyml.Object
	if pre.DiceOverlay != "" {
		if err := json.Unmarshal([]byte(pre.DiceOverlay), &currentDiceObj); err != nil {
			logrus.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
			return oldOverlay, errors.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
		}
		for k, v := range currentDiceObj.Services {
			if _, ok := oldOverlay.Services[k]; ok {
				if oldOverlay.Services[k].Deployments.Replicas != v.Deployments.Replicas {
					oldOverlay.Services[k].Deployments.Replicas = v.Deployments.Replicas
				}
				if oldOverlay.Services[k].Resources.CPU != v.Resources.CPU {
					oldOverlay.Services[k].Resources.CPU = v.Resources.CPU
				}
				if oldOverlay.Services[k].Resources.Mem != v.Resources.Mem {
					oldOverlay.Services[k].Resources.Mem = v.Resources.Mem
				}
			}
		}
	}

	return oldOverlay, nil
}

func genOverlayDataForAudit(oldService *diceyml.Service) *pb.RuntimeInspectServiceDTO {
	return &pb.RuntimeInspectServiceDTO{
		Resources: &pb.Resources{
			Cpu:  oldService.Resources.CPU,
			Mem:  int64(oldService.Resources.Mem),
			Disk: int64(oldService.Resources.Disk),
		},
		Deployments: &pb.Deployments{
			Replicas: uint64(oldService.Deployments.Replicas),
		},
	}
}
