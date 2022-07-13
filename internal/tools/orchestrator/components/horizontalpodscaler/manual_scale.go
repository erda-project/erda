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

package horizontalpodscaler

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/orchestrator/horizontalpodscaler/pb"
	"github.com/erda-project/erda/apistructs"
	patypes "github.com/erda-project/erda/internal/tools/orchestrator/components/horizontalpodscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// processRuntimeScaleRecord 处理单个 runtime 对应的 scale 操作
func (s *hpscalerService) processRuntimeScaleRecord(rsc pb.RuntimeScaleRecord, action string) (*pb.PreDiceDTO, error) {
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
		if rule.IsApplied == patypes.RuntimeHPARuleApplied {
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
	oldServiceReplicas := make(map[string]int)
	for k, v := range rsc.Payload.Services {
		oldService, exists := oldOverlay.Services[k]
		if !exists || oldService == nil {
			oldOverlay.Services[k] = &diceyml.Service{}
		}
		// Local Envs
		if v.Envs != nil {
			oldService.Envs = v.Envs
		}
		// record need update scale's service

		if oldService.Resources.CPU != v.Resources.Cpu || oldService.Resources.Mem != int(v.Resources.Mem) ||
			oldService.Resources.Disk != int(v.Resources.Disk) || oldService.Deployments.Replicas != int(v.Deployments.Replicas) {
			needUpdateServices = append(needUpdateServices, k)
			// record old service's scale for audit
			oldOverlayDataForAudit.Services[k] = genOverlayDataForAudit(oldService)
		} else if action == apistructs.ScaleActionUp {
			// 由于 pre.DiceOverlay 保留了 scale 到 0 之前的副本数（比如2），因而 如果 从 0 恢复到之前非0 副本（2），实际上原来判断是否需要更新的逻辑就认为不需要更新
			needUpdateServices = append(needUpdateServices, k)
			// record old service's scale for audit
			oldOverlayDataForAudit.Services[k] = genOverlayDataForAudit(oldService)
		}

		// 仅更新最新部署的副本数不为 0 的情况，如果，最新部署副本为 0，则保留上次部署时的副本数，用于启动恢复
		if v.Deployments.Replicas == 0 {
			oldServiceReplicas[k] = oldService.Deployments.Replicas
		}

		// Replicas
		oldService.Deployments.Replicas = int(v.Deployments.Replicas)
		// Resources
		oldService.Resources.CPU = v.Resources.Cpu
		oldService.Resources.Mem = int(v.Resources.Mem)
		oldService.Resources.Disk = int(v.Resources.Disk)
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

		// 未设置 action，则需要自行判断用于执行 scale 操作之后的 runtime 状态更新
		if action == "" {
			// 如果所有的 serice 的副本数都调整为0，则是停止操作，Runtime 状态应该为 Stopped，否则不是停止操作，Runtime 状态应该为 Healthy
			action = apistructs.ScaleActionDown
			for _, svc := range sg.Services {
				if svc.Scale > 0 {
					action = apistructs.ScaleActionUp
					break
				}
			}
		}

		sgb, _ := json.Marshal(&sg)
		logrus.Debugf("[processRuntimeScaleRecord] scale service group body is %s", string(sgb))

		if _, err := s.serviceGroupImpl.Scale(sg); err != nil {
			logrus.Errorf("[processRuntimeScaleRecord] process runtime scale failed, ScaleServiceGroup failed for runtime %s for runtime %#v failed, err: %v", uniqueId.Name, uniqueId, err)
			return nil, err
		}

		if action == apistructs.ScaleActionDown || action == apistructs.ScaleActionUp {
			addons, err := s.db.GetUnDeletableAttachMentsByRuntimeID(runtime.ID)
			if err != nil {
				logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
			}
			switch action {
			case apistructs.ScaleActionDown:
				runtime.Status = apistructs.RuntimeStatusStopped
				if err := s.db.UpdateRuntime(runtime); err != nil {
					logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update runtime_status to 'Stopped' in %s for runtime %#v failed, err: %v", runtime.TableName(), uniqueId, err)
				}

				if addons != nil {
					for _, att := range *addons {
						addonRouting, err := s.db.GetInstanceRouting(att.RoutingInstanceID)
						if err != nil {
							logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
							continue
						}

						if addonRouting != nil && addonRouting.PlatformServiceType == 0 {
							att.Deleted = apistructs.AddonScaleDown
							if err := s.db.UpdateAttachment(&att); err != nil {
								logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update table %s for runtime %#v failed, err: %v", att.TableName(), uniqueId, err)
							}
						}
					}
				}

			case apistructs.ScaleActionUp:
				runtime.Status = apistructs.RuntimeStatusHealthy
				if err := s.db.UpdateRuntime(runtime); err != nil {
					logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update runtime_status to 'Healthy' in %s for runtime %#v failed, err: %v", runtime.TableName(), uniqueId, err)
				}

				if addons != nil {
					for _, att := range *addons {
						addonRouting, err := s.db.GetInstanceRouting(att.RoutingInstanceID)
						if err != nil {
							logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update runtime referenced addon attact_count for runtime %#v failed, err: %v", uniqueId, err)
							continue
						}

						if addonRouting != nil && addonRouting.PlatformServiceType == 0 {
							att.Deleted = apistructs.AddonNotDeleted
							if err := s.db.UpdateAttachment(&att); err != nil {
								logrus.Warnf("[processRuntimeScaleRecord] process runtime scale successed, but update table %s for runtime %#v failed, err: %v", att.TableName(), uniqueId, err)
							}
						}
					}
				}
			}
		}
		s.db.UpdateRuntime(runtime)
	}

	// 保留非停止状态下的副本数
	if action != apistructs.ScaleActionDown {
		for k := range oldOverlay.Services {
			if oldOverlay.Services[k].Deployments.Replicas != 0 {
				continue
			}
			replicas, ok := oldServiceReplicas[k]
			if !ok {
				continue
			}
			if replicas > 0 {
				oldOverlay.Services[k].Deployments.Replicas = replicas
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
	}

	logrus.Errorf("[processRuntimeScaleRecord] process runtime scale for runtime %#v successfully", uniqueId)
	return oldOverlayDataForAudit, nil
}

func getPreDeploymentDiceOverlay(pre *dbclient.PreDeployment) (diceyml.Object, error) {
	var oldOverlay diceyml.Object
	if pre.DiceOverlay != "" {
		if err := json.Unmarshal([]byte(pre.DiceOverlay), &oldOverlay); err != nil {
			logrus.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
			return oldOverlay, errors.Errorf("[getPreDeploymentDiceOverlay] Unmarshal PreDeployment record dice_overaly to diceyml.Object for runtime [application_id: %v workspace: %v name: %v ] failed, err: %v", pre.ApplicationId, pre.Workspace, pre.RuntimeName, err)
		}
	} else {
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
