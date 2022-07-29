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
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"modernc.org/mathutil"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
)

func (s *podscalerService) createHPARule(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, serviceRules []*pb.RuntimeServiceHPAConfig) (*pb.CommonResponse, error) {
	mapSVCNameToIndx, sgSvcObject, err := s.getTargetMeta(runtime, serviceRules, nil, true)
	if err != nil {
		return nil, errors.Errorf("[createHPARule] create hpa rule failed, error: %v", err)
	}

	logrus.Infof("[createHPARule] ErdaHPALabelValueCreate Scale return sgHPAObjects: %#v", sgSvcObject)
	for svc, obj := range sgSvcObject {
		idx, ok := mapSVCNameToIndx[svc]
		if !ok {
			continue
		}
		sc := serviceRules[idx].ScaledConfig
		if serviceRules[idx].RuleName != "" {
			sc.RuleName = strings.ToLower(serviceRules[idx].RuleName)
		} else {
			sc.RuleName = svc
		}
		sc.RuleID = uuid.NewString()
		sc.RuntimeID = runtime.ID
		sc.ServiceName = svc
		sc.ApplicationID = appInfo.ID
		sc.OrgID = runtime.OrgID
		sc.RuleNameSpace = obj.Namespace
		sc.ScaleTargetRef = &pb.ScaleTargetRef{
			Kind:       obj.Kind,
			ApiVersion: obj.APIVersion,
			Name:       obj.Name,
		}
		sc.Fallback = &pb.FallBack{
			Replicas: int32(serviceRules[idx].Deployments.Replicas),
		}

		scb, _ := json.Marshal(&sc)
		runtimeSvcHPA := convertRuntimeServiceHPA(userInfo, appInfo, runtime, svc, serviceRules[idx].RuleName, sc.RuleID, obj.Namespace, string(scb))

		applyErr := s.applyOrCancelRule(runtime, runtimeSvcHPA, nil, sc.RuleID, pstypes.ErdaHPALabelValueApply, true)
		if applyErr != nil {
			return nil, errors.Errorf("[applyOrCancelHPARule] applyOrCancelRule failed: %v", applyErr)
		}
		runtimeSvcHPA.IsApplied = pstypes.RuntimePARuleApplied

		err = s.db.CreateHPARule(runtimeSvcHPA)
		if err != nil {
			createErr := errors.Errorf("create hpa rule record failed for runtime %s for service %s, err: %v", runtime.Name, svc, err)
			return nil, errors.New(fmt.Sprintf("[createHPARule] create hpa rule failed, error: %v", createErr))
		}
	}

	return nil, nil
}

func (s *podscalerService) listHPARules(runtime *dbclient.Runtime, services []string) (*pb.ErdaRuntimeHPARules, error) {
	id := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}

	logrus.Infof("[listHPARules] get runtime hpa rules with spec.RuntimeUniqueId: %#v and services [%v]", id, services)
	hpaRules, err := s.db.GetRuntimeHPARulesByServices(id, services)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[listHPARules] get hpa rule failed, error: %v", err))
	}

	rules := make([]*pb.ErdaRuntimeHPARule, 0)
	for _, rule := range hpaRules {
		rules = append(rules, buildRuntimeHPARule(rule))
	}

	return &pb.ErdaRuntimeHPARules{
		RuntimeID: runtime.ID,
		Rules:     rules,
	}, nil
}

func (s *podscalerService) updateHPARules(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, newRulesBase map[string]*pb.ScaledConfig, oldRules map[string]dbclient.RuntimeHPA, req *pb.ErdaRuntimeHPARules) (*pb.CommonResponse, error) {
	for _, rule := range req.Rules {
		if rule.ScaledConfig == nil {
			return nil, errors.Errorf("[updateHPARules] update hpa rule failed for svc %s: scaledConfig not set", rule.ServiceName)
		}

		ruleHPA := oldRules[rule.RuleID]
		newRule := newRulesBase[rule.RuleID]
		needUpdate := false

		if rule.ScaledConfig.MinReplicaCount >= 0 && newRule.MinReplicaCount != rule.ScaledConfig.MinReplicaCount {
			needUpdate = true
			newRule.MinReplicaCount = rule.ScaledConfig.MinReplicaCount
		}

		if rule.ScaledConfig.MaxReplicaCount > 0 && newRule.MaxReplicaCount != rule.ScaledConfig.MaxReplicaCount {
			needUpdate = true
			newRule.MaxReplicaCount = rule.ScaledConfig.MaxReplicaCount
		}

		if rule.ScaledConfig.PollingInterval != newRule.PollingInterval {
			needUpdate = true
			newRule.PollingInterval = rule.ScaledConfig.PollingInterval
		}

		if rule.ScaledConfig.CooldownPeriod != newRule.CooldownPeriod {
			needUpdate = true
			newRule.CooldownPeriod = rule.ScaledConfig.CooldownPeriod
		}

		if rule.ScaledConfig.Advanced != nil && rule.ScaledConfig.Advanced.RestoreToOriginalReplicaCount != newRule.Advanced.RestoreToOriginalReplicaCount {
			needUpdate = true
			newRule.Advanced.RestoreToOriginalReplicaCount = rule.ScaledConfig.Advanced.RestoreToOriginalReplicaCount
		}

		if !isTriggersEqual(rule.ScaledConfig.Triggers, newRule.Triggers) {
			needUpdate = true
			newRule.Triggers = rule.ScaledConfig.Triggers
		}

		if rule.ScaledConfig.Fallback != nil {
			if (rule.ScaledConfig.Fallback.FailureThreshold != newRule.Fallback.FailureThreshold) && rule.ScaledConfig.Fallback.FailureThreshold > 0 {
				needUpdate = true
				newRule.Fallback.FailureThreshold = rule.ScaledConfig.Fallback.FailureThreshold
			}
			if (rule.ScaledConfig.Fallback.Replicas != newRule.Fallback.Replicas) && rule.ScaledConfig.Fallback.Replicas > 0 {
				needUpdate = true
				newRule.Fallback.Replicas = rule.ScaledConfig.Fallback.Replicas
			}
		}

		if needUpdate {
			newRulesStr, _ := json.Marshal(*newRule)
			updatedRule := &dbclient.RuntimeHPA{
				ID:                     ruleHPA.ID,
				RuleName:               ruleHPA.RuleName,
				RuleNameSpace:          ruleHPA.RuleNameSpace,
				OrgID:                  ruleHPA.OrgID,
				OrgName:                ruleHPA.OrgName,
				OrgDisPlayName:         ruleHPA.OrgDisPlayName,
				ProjectID:              ruleHPA.ProjectID,
				ProjectName:            ruleHPA.ProjectName,
				ProjectDisplayName:     ruleHPA.ProjectDisplayName,
				ApplicationID:          ruleHPA.ApplicationID,
				ApplicationName:        ruleHPA.ApplicationName,
				ApplicationDisPlayName: ruleHPA.ApplicationDisPlayName,
				RuntimeID:              ruleHPA.RuntimeID,
				RuntimeName:            ruleHPA.RuntimeName,
				ClusterName:            ruleHPA.ClusterName,
				Workspace:              ruleHPA.Workspace,
				UserID:                 userInfo.ID,
				UserName:               userInfo.Name,
				NickName:               userInfo.Nick,
				ServiceName:            ruleHPA.ServiceName,
				Rules:                  string(newRulesStr),
				IsApplied:              ruleHPA.IsApplied,
			}

			if ruleHPA.IsApplied == pstypes.RuntimePARuleApplied {
				// 已部署，需要删除，然后重新部署
				reApplyErr := s.applyOrCancelRule(runtime, updatedRule, nil, updatedRule.ID, pstypes.ErdaHPALabelValueReApply, true)
				if reApplyErr != nil {
					return nil, errors.Errorf("[updateHPARules] applyOrCancelRule failed: %v", reApplyErr)
				}
			}

			// 未部署，直接更新
			err := s.db.UpdateHPARule(updatedRule)
			if err != nil {
				return nil, errors.Errorf("[updateHPARules] update hpa rule failed for svc %s: update rule by rule id %s with error: %v", rule.ServiceName, rule.RuleID, err)
			}
		}
	}
	return nil, nil
}

func (s *podscalerService) deleteHPARule(userID string, runtime *dbclient.Runtime, ruleIds []string) (*pb.CommonResponse, error) {

	var err error
	ruleIdsMap := make(map[string]dbclient.RuntimeHPA)
	if len(ruleIds) == 0 {
		//delete all rules in a runtime
		rules, err := s.db.GetRuntimeHPARulesByRuntimeId(runtime.ID)
		if err != nil {
			return nil, errors.Errorf("[deleteHPARule] GetErdaRuntimeHPARulesByRuntimeId failed: %v", err)
		}

		for _, rule := range rules {
			ruleIds = append(ruleIds, rule.ID)
			ruleIdsMap[rule.ID] = rule
		}
	}

	for _, ruleId := range ruleIds {
		var runtimeHPA dbclient.RuntimeHPA
		rule, ok := ruleIdsMap[ruleId]
		if ok {
			runtimeHPA = rule
		} else {
			runtimeHPA, err = s.db.GetRuntimeHPARuleByRuleId(ruleId)
			if err != nil {
				return nil, errors.Errorf("[deleteHPARule] GetErdaHRuntimePARuleByRuleId failed: %v", err)
			}
		}

		if runtimeHPA.IsApplied == pstypes.RuntimePARuleApplied {
			// 已部署，需要删除
			cancelErr := s.applyOrCancelRule(runtime, &runtimeHPA, nil, runtimeHPA.ID, pstypes.ErdaHPALabelValueCancel, true)
			if cancelErr != nil {
				return nil, errors.Errorf("[deleteHPARule] applyOrCancelRule failed: %v", cancelErr)
			}
		}

		if err = s.db.DeleteRuntimeHPAEventsByRuleId(ruleId); err != nil {
			logrus.Warnf("[deleteHPARule] DeleteErdaRuntimeHPAEventsByRuleId failed: %v", err)
		}

		if err = s.db.DeleteRuntimeHPARulesByRuleId(ruleId); err != nil {
			return nil, errors.Errorf("[deleteHPARule] DeleteErdaHRuntimePARulesByRuleId failed: %v", err)
		}
	}

	return nil, nil
}

func (s *podscalerService) applyOrCancelHPARule(userInfo *apistructs.UserInfo, runtime *dbclient.Runtime, RuleAction []*pb.RuleAction) (*pb.CommonResponse, error) {
	for idx := range RuleAction {
		hpaRule, err := s.db.GetRuntimeHPARuleByRuleId(RuleAction[idx].RuleId)
		if err != nil {
			return nil, errors.Errorf("[applyOrCancelHPARule] GetErdaRuntimeHPARuleByRuleId failed: %v", err)
		}

		switch RuleAction[idx].Action {
		case pstypes.ErdaPARuleActionApply:
			if hpaRule.IsApplied == pstypes.RuntimePARuleCanceled {
				// 未部署，需要部署
				applyErr := s.applyOrCancelRule(runtime, &hpaRule, nil, RuleAction[idx].RuleId, pstypes.ErdaHPALabelValueApply, true)
				if applyErr != nil {
					return nil, errors.Errorf("[applyOrCancelHPARule] applyOrCancelRule failed: %v", applyErr)
				}
				hpaRule.UserID = userInfo.ID
				hpaRule.UserName = userInfo.Name
				hpaRule.NickName = userInfo.Nick
				hpaRule.IsApplied = pstypes.RuntimePARuleApplied
				err = s.db.UpdateHPARule(&hpaRule)
				if err != nil {
					return nil, errors.Errorf("[applyOrCancelHPARule] update rule with ruleId %s error: %v", hpaRule.ID, err)
				}
			} else {
				// 已部署，无需部署
				return nil, errors.Errorf("[applyOrCancelHPARule] hpa rule %v have applied, no need apply it again", hpaRule.ID)
			}

		case pstypes.ErdaPARuleActionCancel:
			if hpaRule.IsApplied == pstypes.RuntimePARuleApplied {
				// 未删除，需要删除
				cancelErr := s.applyOrCancelRule(runtime, nil, nil, RuleAction[idx].RuleId, pstypes.ErdaHPALabelValueCancel, true)
				if cancelErr != nil {
					return nil, errors.Errorf("[applyOrCancelHPARule] update rule with ruleId %s for applyOrCancelRule error: %v", hpaRule.ID, cancelErr)
				}
				hpaRule.UserID = userInfo.ID
				hpaRule.UserName = userInfo.Name
				hpaRule.NickName = userInfo.Nick
				hpaRule.IsApplied = pstypes.RuntimePARuleCanceled
				err = s.db.UpdateHPARule(&hpaRule)
				if err != nil {
					return nil, errors.Errorf("[applyOrCancelHPARule] UpdateErdaHPARule update rule with ruleId %s error: %v", hpaRule.ID, err)
				}
			} else {
				// 已删除，无需删除
				return nil, errors.Errorf("[applyOrCancelHPARule] hpa rule id %v have canceled, no need cancel it again", hpaRule.ID)
			}

		default:
			return nil, errors.Errorf("[applyOrCancelHPARule] unknown action: %s", RuleAction[idx].Action)
		}
	}

	return nil, nil
}

func (s *podscalerService) listHPAEvents(runtimeId uint64, services []string) (*pb.ErdaRuntimeHPAEvents, error) {
	hpaEvents, err := s.db.GetRuntimeHPAEventsByServices(runtimeId, services)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[listHPAEvents] get hpa events for runtimeId [%d] for service [%v] failed, error: %v", runtimeId, services, err))
	}

	resultEvents, err := convertHPAEventInfoToErdaRuntimeHPAEvent(hpaEvents)
	if err != nil {
		return nil, err
	}

	return &pb.ErdaRuntimeHPAEvents{
		Events: resultEvents,
	}, nil
}

func isTriggersEqual(old, new []*pb.ScaleTriggers) bool {
	if len(old) != len(new) {
		return false
	}

	for _, vold := range old {
		oldInNew := false
		for _, vnew := range new {
			if vnew.Type == vold.Type {
				oldInNew = true
				if !reflect.DeepEqual(vnew.Metadata, vold.Metadata) {
					return false
				}
				break
			}
		}
		if !oldInNew {
			return false
		}
	}
	return true
}

func convertRuntimeServiceHPA(userInfo *apistructs.UserInfo, appInfo *apistructs.ApplicationDTO, runtime *dbclient.Runtime, serviceName, ruleName, ruleID, namespace, rulesJson string) *dbclient.RuntimeHPA {
	if ruleName == "" {
		ruleName = serviceName
	}
	return &dbclient.RuntimeHPA{
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

func buildRuntimeHPARule(rule dbclient.RuntimeHPA) *pb.ErdaRuntimeHPARule {
	uid, _ := strconv.ParseUint(rule.UserID, 10, 64)
	scaledConfig := pb.ScaledConfig{}
	json.Unmarshal([]byte(rule.Rules), &scaledConfig)

	return &pb.ErdaRuntimeHPARule{
		RuleID:      rule.ID,
		CreateAt:    timestamppb.New(rule.CreatedAt),
		UpdateAt:    timestamppb.New(rule.UpdatedAt),
		ServiceName: rule.ServiceName,
		RuleName:    rule.RuleName,
		UserInfo: &pb.UserInfo{
			UserID:       uid,
			UserName:     rule.UserName,
			UserNickName: rule.NickName,
		},
		ScaledConfig: &scaledConfig,
		IsApplied:    rule.IsApplied,
	}
}

func validateHPARuleConfigCustom(serviceName string, maxReplicas int32, scaledConf *pb.ScaledConfig) error {

	if scaledConf == nil {
		return errors.Errorf("service %s not set scaledConfig", serviceName)
	}

	if scaledConf.MinReplicaCount < 0 {
		return errors.Errorf("service %s set invalid scaledConfig.minReplicaCount", serviceName)
	}

	if scaledConf.MaxReplicaCount <= 0 || scaledConf.MinReplicaCount > scaledConf.MaxReplicaCount {
		return errors.Errorf("service %s not set invalid scaledConfig.maxReplicaCount", serviceName)
	}

	if scaledConf.MaxReplicaCount > maxReplicas {
		return errors.Errorf("service %s set invalid scaledConfig.maxReplicaCount, need  <=%d", serviceName, maxReplicas)
	}

	if len(scaledConf.Triggers) == 0 {
		return errors.Errorf("service %s not set scaledConfig.triggers", serviceName)
	}

	for idx, trigger := range scaledConf.Triggers {
		if trigger.Type != pstypes.ErdaHPATriggerCron && trigger.Type != pstypes.ErdaHPATriggerCPU && trigger.Type != pstypes.ErdaHPATriggerMemory && trigger.Type != pstypes.ErdaHPATriggerExternal {
			return errors.Errorf("service %s with scaledConfig.triggers[%d] with unsupport trigger type %s", serviceName, idx, trigger.Type)
		}
		if len(trigger.Metadata) == 0 {
			return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata", serviceName, idx, trigger.Type)
		}

		switch trigger.Type {
		case pstypes.ErdaHPATriggerCPU, pstypes.ErdaHPATriggerMemory:
			val, ok := trigger.Metadata["type"]
			if !ok || val != pstypes.ErdaHPATriggerCPUMetaType {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata type or invalid type value (need value 'Utilization')", serviceName, idx, trigger.Type)
			}
			_, ok = trigger.Metadata["value"]
			if !ok {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata value", serviceName, idx, trigger.Type)
			}

			value, err := strconv.Atoi(trigger.Metadata["value"])
			if err != nil || value >= 100 || value <= 0 {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata value in range ( 0 < value < 100)", serviceName, idx, trigger.Type)
			}

		case pstypes.ErdaHPATriggerCron:
			val, ok := trigger.Metadata[pstypes.ErdaHPATriggerCronMetaTimeZone]
			if !ok || val == "" {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata %s or invalid value (need as cron express as 'Asia/Shanghai')", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaTimeZone)
			}

			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			val, ok = trigger.Metadata[pstypes.ErdaHPATriggerCronMetaStart]
			if !ok || val == "" {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata %s or invalid value (need as cron express as '30 * * * *')", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaStart)
			}

			_, err := parser.Parse(val)
			if err != nil {
				return fmt.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s  error parsing start schedule value [%s]: %s", serviceName, idx, trigger.Type, val, err)
			}

			val, ok = trigger.Metadata[pstypes.ErdaHPATriggerCronMetaEnd]
			if !ok || val == "" {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata %s or invalid value (need as cron express as '30 * * * *')", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaEnd)
			}

			_, err = parser.Parse(val)
			if err != nil {
				return fmt.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s  error parsing end schedule value [%s]: %s", serviceName, idx, trigger.Type, val, err)
			}

			val, ok = trigger.Metadata[pstypes.ErdaHPATriggerCronMetaDesiredReplicas]
			if !ok || val == "" {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata %s ", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaDesiredReplicas)
			}

			replicas, err := strconv.Atoi(val)
			if err != nil {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set medatdata %s or invalid value (need as integer number)", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaDesiredReplicas)
			}
			if int32(replicas) > scaledConf.MaxReplicaCount {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s set medatdata %s need <= maxReplicaCount(%d), ", serviceName, idx, trigger.Type, pstypes.ErdaHPATriggerCronMetaDesiredReplicas, scaledConf.MaxReplicaCount)
			}

		default:
			if len(trigger.Metadata) < 2 {
				return errors.Errorf("service %s with scaledConfig.triggers[%d] with trigger type %s not set enough medatdata", serviceName, idx, trigger.Type)
			}
		}
	}
	return nil
}

func validateHPARuleConfig(serviceRules []*pb.RuntimeServiceHPAConfig) error {
	for idx, rule := range serviceRules {
		if rule.Deployments == nil {
			return errors.Errorf("service %s not set deployments", rule.RuleName)
		}

		if rule.Deployments.Replicas <= 0 {
			return errors.Errorf("service %s with invalid setdeployments.replicas %v", rule.RuleName, rule.Deployments.Replicas)
		}

		maxReplicas := mathutil.MinInt32(10*int32(rule.Deployments.Replicas), pstypes.ErdaHPADefaultMaxReplicaCount)

		if rule.Resources == nil {
			return errors.Errorf("service %s not set resources", rule.RuleName)
		}
		if rule.Resources.Mem <= 0 || rule.Resources.Cpu <= float64(0) {
			return errors.Errorf("service %s set resources with invalid cpu or mem", rule.RuleName)
		}

		if rule.ScaledConfig == nil {
			return errors.Errorf("service %s not set scaledConfig", rule.RuleName)
		}

		if rule.ScaledConfig.Advanced == nil {
			serviceRules[idx].ScaledConfig.Advanced = &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			}
		}

		err := validateHPARuleConfigCustom(rule.RuleName, maxReplicas, rule.ScaledConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertHPAEventInfoToErdaRuntimeHPAEvent(hpaEvents []dbclient.HPAEventInfo) ([]*pb.ErdaRuntimeHPAEvent, error) {
	result := make([]*pb.ErdaRuntimeHPAEvent, 0)
	for _, ev := range hpaEvents {

		evInfo := pstypes.EventDetail{}
		err := json.Unmarshal([]byte(ev.Event), &evInfo)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[listHPAEvents] Unmarshal hpa events for runtimeId [%d] for service [%v] error: %v", ev.RuntimeID, ev.ServiceName, err))
		}

		result = append(result, &pb.ErdaRuntimeHPAEvent{
			ServiceName: ev.ServiceName,
			RuleId:      ev.ID,
			Event: &pb.HPAEventDetail{
				CreateAt:     timestamppb.New(ev.CreatedAt),
				Type:         evInfo.Type,
				Reason:       evInfo.Reason,
				EventMessage: evInfo.Message,
			},
		})
	}
	return result, nil
}
