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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"modernc.org/mathutil"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
)

func (s *podscalerService) createVPARule(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, serviceRules []*pb.RuntimeServiceVPAConfig) (*pb.CommonResponse, error) {
	mapSVCNameToIndx := make(map[string]int)
	mapSVCNameToIndx, sgSvcObject, err := s.getTargetMeta(runtime, nil, serviceRules, false)
	if err != nil {
		return nil, errors.Errorf("[createVPARule] create hpa rule failed, error: %v", err)
	}

	logrus.Infof("ErdaHPALabelValueCreate Scale return sgHPAObjects: %#v", sgSvcObject)
	for svc, obj := range sgSvcObject {
		idx, ok := mapSVCNameToIndx[svc]
		if !ok {
			continue
		}

		ruleName := svc
		if serviceRules[idx].RuleName != "" {
			ruleName = strings.ToLower(serviceRules[idx].RuleName)
		}

		updateMode := pstypes.ErdaVPAUpdaterModeAuto
		if serviceRules[idx].UpdateMode != "" {
			updateMode = serviceRules[idx].UpdateMode
		}

		ruleID := uuid.NewString()
		vpa := &pb.RuntimeServiceVPAConfig{
			RuleID:        ruleID,
			RuleName:      ruleName,
			RuntimeID:     runtime.ID,
			ApplicationID: appInfo.ID,
			ProjectID:     runtime.ProjectID,
			OrgID:         runtime.OrgID,
			ServiceName:   svc,
			ScaleTargetRef: &pb.ScaleTargetRef{
				Kind:       obj.Kind,
				ApiVersion: obj.APIVersion,
				Name:       obj.Name,
			},
			RuleNameSpace: obj.Namespace,
			UpdateMode:    updateMode,
			Deployments:   serviceRules[idx].Deployments,
			Resources:     serviceRules[idx].Resources,
			MaxResources:  serviceRules[idx].MaxResources,
		}

		vpab, _ := json.Marshal(&vpa)
		runtimeSvcVPA := convertRuntimeServiceVPA(userInfo, appInfo, runtime, svc, serviceRules[idx].RuleName, ruleID, obj.Namespace, string(vpab))

		applyErr := s.applyOrCancelRule(runtime, nil, runtimeSvcVPA, ruleID, pstypes.ErdaVPALabelValueApply, false)
		if applyErr != nil {
			return nil, errors.Errorf("[applyOrCancelHPARule] applyOrCancelRule failed: %v", applyErr)
		}
		runtimeSvcVPA.IsApplied = pstypes.RuntimePARuleApplied

		err = s.db.CreateVPARule(runtimeSvcVPA)
		if err != nil {
			createErr := errors.Errorf("create vpa rule record failed for runtime %s for service %s, err: %v", runtime.Name, svc, err)
			return nil, errors.New(fmt.Sprintf("[createHPARule] create vpa rule failed, error: %v", createErr))
		}
	}

	return nil, nil
}

func (s *podscalerService) deleteVPARule(userID string, runtime *dbclient.Runtime, ruleIds []string) (*pb.CommonResponse, error) {
	var err error
	ruleIdsMap := make(map[string]dbclient.RuntimeVPA)
	if len(ruleIds) == 0 {
		//delete all rules in a runtime
		rules, err := s.db.GetRuntimeVPARulesByRuntimeId(runtime.ID)
		if err != nil {
			return nil, errors.Errorf("[deleteVPARule] GetErdaRuntimePARulesByRuntimeId failed: %v", err)
		}

		for _, rule := range rules {
			ruleIds = append(ruleIds, rule.ID)
			ruleIdsMap[rule.ID] = rule
		}
	}

	for _, ruleId := range ruleIds {
		var runtimeVPA dbclient.RuntimeVPA
		rule, ok := ruleIdsMap[ruleId]
		if ok {
			runtimeVPA = rule
		} else {
			runtimeVPA, err = s.db.GetRuntimeVPARuleByRuleId(ruleId)
			if err != nil {
				return nil, errors.Errorf("[deleteVPARule] GetErdaHRuntimePARuleByRuleId failed: %v", err)
			}
		}

		if runtimeVPA.IsApplied == pstypes.RuntimePARuleApplied {
			// 已部署，需要删除
			cancelErr := s.applyOrCancelRule(runtime, nil, &runtimeVPA, runtimeVPA.ID, pstypes.ErdaVPALabelValueCancel, false)
			if cancelErr != nil {
				return nil, errors.Errorf("[deleteVPARule] applyOrCancelRule failed: %v", cancelErr)
			}
		}

		if err = s.db.DeleteRuntimeVPARulesByRuleId(ruleId); err != nil {
			return nil, errors.Errorf("[deleteHPARule] DeleteErdaHRuntimePARulesByRuleId failed: %v", err)
		}
	}

	return nil, nil
}

func (s *podscalerService) listVPARules(runtime *dbclient.Runtime, services []string) (*pb.ErdaRuntimeVPARules, error) {
	id := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}

	logrus.Infof("[listVPARules] get runtime vpa rules with spec.RuntimeUniqueId: %#v and services [%v]", id, services)
	hpaRules, err := s.db.GetRuntimeVPARulesByServices(id, services)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[listVPARules] get vpa rule failed, error: %v", err))
	}

	rules := make([]*pb.ErdaRuntimeVPARule, 0)
	for _, rule := range hpaRules {
		rules = append(rules, buildRuntimeVPARule(rule))
	}

	return &pb.ErdaRuntimeVPARules{
		RuntimeID: runtime.ID,
		Rules:     rules,
	}, nil
}

func buildRuntimeVPARule(rule dbclient.RuntimeVPA) *pb.ErdaRuntimeVPARule {
	uid, _ := strconv.ParseUint(rule.UserID, 10, 64)
	scaledConfig := pb.RuntimeServiceVPAConfig{}
	json.Unmarshal([]byte(rule.Rules), &scaledConfig)

	return &pb.ErdaRuntimeVPARule{
		//RuleID:      rule.ID,
		CreateAt:    timestamppb.New(rule.CreatedAt),
		UpdateAt:    timestamppb.New(rule.UpdatedAt),
		ServiceName: rule.ServiceName,
		//RuleName:    rule.RuleName,
		UserInfo: &pb.UserInfo{
			UserID:       uid,
			UserName:     rule.UserName,
			UserNickName: rule.NickName,
		},
		Rule: &pb.RuntimeServiceVPAConfig{
			RuleID:         rule.ID,
			RuleName:       rule.RuleName,
			RuntimeID:      rule.RuntimeID,
			ApplicationID:  rule.ApplicationID,
			ProjectID:      rule.ProjectID,
			OrgID:          rule.OrgID,
			ServiceName:    rule.ServiceName,
			ScaleTargetRef: scaledConfig.ScaleTargetRef,
			RuleNameSpace:  rule.RuleNameSpace,
			UpdateMode:     scaledConfig.UpdateMode,
			Deployments:    scaledConfig.Deployments,
			Resources:      scaledConfig.Resources,
			MaxResources:   scaledConfig.MaxResources,
		},
		IsApplied: rule.IsApplied,
	}
}

func (s *podscalerService) updateVPARules(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, newRules map[string]*pb.RuntimeServiceVPAConfig, oldRules map[string]*pb.RuntimeServiceVPAConfig, oldVPAs map[string]dbclient.RuntimeVPA, req *pb.ErdaRuntimeHPARules) (*pb.CommonResponse, error) {
	for id, newRule := range newRules {
		oldRule, ok := oldRules[id]
		if !ok {
			continue
		}

		oldVPA, ok := oldVPAs[id]
		if !ok {
			continue
		}

		needUpdate := false

		if newRule.UpdateMode != oldRule.UpdateMode {
			oldRule.UpdateMode = newRule.UpdateMode
			needUpdate = true
		}

		if newRule.MaxResources != nil {
			if newRule.MaxResources.Cpu != oldRule.MaxResources.Cpu {
				oldRule.MaxResources.Cpu = newRule.MaxResources.Cpu
				needUpdate = true
			}

			if newRule.MaxResources.Mem != oldRule.MaxResources.Mem {
				oldRule.MaxResources.Mem = newRule.MaxResources.Mem
				needUpdate = true
			}

		}

		if needUpdate {
			newRulesStr, _ := json.Marshal(*oldRule)
			updatedRule := &dbclient.RuntimeVPA{
				ID:                     oldVPA.ID,
				RuleName:               oldVPA.RuleName,
				RuleNameSpace:          oldVPA.RuleNameSpace,
				OrgID:                  oldVPA.OrgID,
				OrgName:                oldVPA.OrgName,
				OrgDisPlayName:         oldVPA.OrgDisPlayName,
				ProjectID:              oldVPA.ProjectID,
				ProjectName:            oldVPA.ProjectName,
				ProjectDisplayName:     oldVPA.ProjectDisplayName,
				ApplicationID:          oldRule.ApplicationID,
				ApplicationName:        oldVPA.ApplicationName,
				ApplicationDisPlayName: oldVPA.ApplicationDisPlayName,
				RuntimeID:              oldVPA.RuntimeID,
				RuntimeName:            oldVPA.RuntimeName,
				ClusterName:            oldVPA.ClusterName,
				Workspace:              oldVPA.Workspace,
				UserID:                 userInfo.ID,
				UserName:               userInfo.Name,
				NickName:               userInfo.Nick,
				ServiceName:            oldVPA.ServiceName,
				Rules:                  string(newRulesStr),
				IsApplied:              oldVPA.IsApplied,
			}

			if oldVPA.IsApplied == pstypes.RuntimePARuleApplied {
				// 已部署，需要删除，然后重新部署
				reApplyErr := s.applyOrCancelRule(runtime, nil, updatedRule, oldVPA.ID, pstypes.ErdaVPALabelValueReApply, false)
				if reApplyErr != nil {
					return nil, errors.Errorf("[updateVPARules] applyOrCancelRule failed: %v", reApplyErr)
				}
			}

			// 未部署，直接更新
			err := s.db.UpdateVPARule(updatedRule)
			if err != nil {
				return nil, errors.Errorf("[updateVPARules] update vpa rule failed for svc %s: update rule by rule id %s with error: %v", oldVPA.ServiceName, oldVPA.ID, err)
			}
		}
	}
	return nil, nil
}

func (s *podscalerService) applyOrCancelVPARule(userInfo *apistructs.UserInfo, runtime *dbclient.Runtime, RuleAction []*pb.RuleAction) (*pb.CommonResponse, error) {
	for idx := range RuleAction {
		vpaRule, err := s.db.GetRuntimeVPARuleByRuleId(RuleAction[idx].RuleId)
		if err != nil {
			return nil, errors.Errorf("[applyOrCancelVPARule] GetErdaRuntimeVPARuleByRuleId failed: %v", err)
		}

		switch RuleAction[idx].Action {
		case pstypes.ErdaPARuleActionApply:
			if vpaRule.IsApplied == pstypes.RuntimePARuleCanceled {
				// 未部署，需要部署
				applyErr := s.applyOrCancelRule(runtime, nil, &vpaRule, RuleAction[idx].RuleId, pstypes.ErdaVPALabelValueApply, false)
				if applyErr != nil {
					return nil, errors.Errorf("[applyOrCancelVPARule] applyOrCancelRule vpa failed: %v", applyErr)
				}
				vpaRule.UserID = userInfo.ID
				vpaRule.UserName = userInfo.Name
				vpaRule.NickName = userInfo.Nick
				vpaRule.IsApplied = pstypes.RuntimePARuleApplied
				err = s.db.UpdateVPARule(&vpaRule)
				if err != nil {
					return nil, errors.Errorf("[applyOrCancelVPARule] update rule with ruleId %s error: %v", vpaRule.ID, err)
				}
			} else {
				// 已部署，无需部署
				return nil, errors.Errorf("[applyOrCancelVPARule] vpa rule %v have applied, no need apply it again", vpaRule.ID)
			}

		case pstypes.ErdaPARuleActionCancel:
			if vpaRule.IsApplied == pstypes.RuntimePARuleApplied {
				// 未删除，需要删除
				cancelErr := s.applyOrCancelRule(runtime, nil, nil, RuleAction[idx].RuleId, pstypes.ErdaVPALabelValueCancel, false)
				if cancelErr != nil {
					return nil, errors.Errorf("[applyOrCancelVPARule] update rule with ruleId %s for applyOrCancelRule error: %v", vpaRule.ID, cancelErr)
				}
				vpaRule.UserID = userInfo.ID
				vpaRule.UserName = userInfo.Name
				vpaRule.NickName = userInfo.Nick
				vpaRule.IsApplied = pstypes.RuntimePARuleCanceled
				err = s.db.UpdateVPARule(&vpaRule)
				if err != nil {
					return nil, errors.Errorf("[applyOrCancelVPARule] UpdateErdaVPARule update vpa rule with ruleId %s error: %v", vpaRule.ID, err)
				}
			} else {
				// 已删除，无需删除
				return nil, errors.Errorf("[applyOrCancelVPARule] vpa rule id %v have canceled, no need cancel it again", vpaRule.ID)
			}

		default:
			return nil, errors.Errorf("[applyOrCancelVPARule] unknown action: %s", RuleAction[idx].Action)
		}
	}

	return nil, nil
}

func validVPAMaxResources(resources *pb.Resources, maxCPU float64, maxMemory int64) error {
	if resources.Cpu > maxCPU || resources.Cpu < pstypes.ErdaVPAMinResourceCPU {
		return errors.Errorf("vpa maxResources.Cpu must in range [%v, %v], but has value %v", pstypes.ErdaVPAMinResourceCPU, maxCPU, resources.Cpu)
	}

	if resources.Mem > maxMemory || resources.Mem < pstypes.ErdaVPAMinResourceMemory {
		return errors.Errorf("vpa maxResources.Mem must in range [%v, %v], but has value %v", pstypes.ErdaVPAMinResourceMemory, maxMemory, resources.Mem)
	}
	return nil
}

func validateVPARuleCreateConfig(serviceRules []*pb.RuntimeServiceVPAConfig, isUpdate bool) error {
	for idx, rule := range serviceRules {
		if !isUpdate && rule.ServiceName == "" {
			return errors.Errorf("serviceName not set")
		}

		switch rule.UpdateMode {
		case pstypes.ErdaVPAUpdaterModeAuto, "":
		case "Recreate":
		case "Initial":
		case "Off":
		default:
			return errors.Errorf("unknown vpa updateMode: %s", rule.UpdateMode)
		}

		if isUpdate {
			if rule.RuleID == "" {
				return errors.Errorf("not set vpa rule id")
			}

			if rule.MaxResources == nil {
				return errors.Errorf("not set vpa maxResources")
			}

			err := validVPAMaxResources(rule.MaxResources, pstypes.ErdaVPAMaxResourceCPU, pstypes.ErdaVPAMaxResourceMemory)
			if err != nil {
				return err
			}
		} else {
			maxCPU := pstypes.ErdaVPADefaultMaxCPU
			maxMemory := pstypes.ErdaVPADefaultMaxMemory

			if rule.Resources != nil {
				if rule.Resources.Cpu > pstypes.ErdaVPADefaultMaxCPU {
					maxCPU = math.Min(pstypes.ErdaVPADefaultResourceMinFactor*rule.Resources.Cpu, pstypes.ErdaVPAMaxResourceCPU)
				} else {
					maxCPU = math.Min(pstypes.ErdaVPADefaultResourceMaxFactor*rule.Resources.Cpu, pstypes.ErdaVPADefaultResourceMinFactor*pstypes.ErdaVPADefaultMaxCPU)
				}

				if rule.Resources.Mem > pstypes.ErdaVPADefaultMaxMemory {
					maxMemory = mathutil.MinInt64(pstypes.ErdaVPADefaultResourceMinFactor*rule.Resources.Mem, pstypes.ErdaVPAMaxResourceMemory)
				} else {
					maxMemory = mathutil.MinInt64(pstypes.ErdaVPADefaultResourceMaxFactor*rule.Resources.Mem, pstypes.ErdaVPADefaultResourceMinFactor*pstypes.ErdaVPADefaultMaxMemory)
				}
			}

			if rule.MaxResources != nil {
				err := validVPAMaxResources(rule.MaxResources, maxCPU, maxMemory)
				if err != nil {
					return err
				}
				continue
			} else {
				serviceRules[idx].MaxResources = &pb.Resources{
					Cpu: maxCPU,
					Mem: maxMemory,
				}
			}

			err := validVPAMaxResources(rule.MaxResources, maxCPU, maxMemory)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertRuntimeServiceVPA(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, serviceName, ruleName, ruleID, namespace, rulesJson string) *dbclient.RuntimeVPA {
	if ruleName == "" {
		ruleName = serviceName
	}
	return &dbclient.RuntimeVPA{
		ID:                     ruleID,
		RuleName:               ruleName,
		RuleNameSpace:          namespace,
		OrgID:                  appInfo.OrgID,
		OrgName:                appInfo.OrgName,
		OrgDisPlayName:         appInfo.OrgDisplayName,
		ProjectID:              appInfo.ProjectID,
		ProjectName:            appInfo.ProjectName,
		ProjectDisplayName:     appInfo.ProjectDisplayName,
		ApplicationID:          appInfo.ID,
		ApplicationName:        appInfo.Name,
		ApplicationDisPlayName: appInfo.DisplayName,
		RuntimeID:              runtime.ID,
		RuntimeName:            runtime.Name,
		ClusterName:            runtime.ClusterName,
		Workspace:              runtime.Workspace,
		UserID:                 userInfo.ID,
		UserName:               userInfo.Name,
		NickName:               userInfo.Nick,
		ServiceName:            serviceName,
		Rules:                  rulesJson,
		IsApplied:              pstypes.RuntimePARuleCanceled,
	}
}

func (s *podscalerService) listVPAServiceRecommendations(runtimeId uint64, services []string) (*pb.ErdaRuntimeVPARecommendations, error) {

	logrus.Infof("[listVPAServiceRecommendations] get runtime vpa recommendations for runtimeId: %v and services [%v]", runtimeId, services)
	vpaRecomms, err := s.db.GetRuntimeVPARecommendationsByServices(runtimeId, services)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[listVPAServiceRecommendations] get vpa recommendations failed, error: %v", err))
	}

	rs := make([]*pb.ErdaServiceRecommendation, 0)
	for _, vpaRecommendation := range vpaRecomms {
		rs = append(rs, buildRuntimeVPARecommendation(vpaRecommendation))
	}

	return &pb.ErdaRuntimeVPARecommendations{
		RuntimeID:              runtimeId,
		ServiceRecommendations: rs,
	}, nil
}

func buildRuntimeVPARecommendation(rule dbclient.RuntimeVPAContainerRecommendation) *pb.ErdaServiceRecommendation {
	return &pb.ErdaServiceRecommendation{
		Id:          rule.ID,
		RuleID:      rule.RuleID,
		RuleName:    rule.RuleName,
		CreateAt:    timestamppb.New(rule.CreatedAt),
		ServiceName: rule.ServiceName,
		ContainerRecommendation: &pb.VPAContainerRecommendation{
			ContainerName: rule.ContainerName,
			LowerBound: &pb.Resources{
				Cpu: rule.LowerCPURequest,
				Mem: int64(rule.LowerMemoryRequest),
			},
			UpperBound: &pb.Resources{
				Cpu: rule.UpperCPURequest,
				Mem: int64(rule.UpperMemoryRequest),
			},
			Target: &pb.Resources{
				Cpu: rule.TargetCPURequest,
				Mem: int64(rule.TargetMemoryRequest),
			},
			UncappedTarget: &pb.Resources{
				Cpu: rule.UncappedCPURequest,
				Mem: int64(rule.UncappedMemoryRequest),
			},
		},
	}
}
