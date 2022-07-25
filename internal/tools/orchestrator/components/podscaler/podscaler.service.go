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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

type podscalerService struct {
	//p *provider
	bundle           BundleService
	db               DBService
	serviceGroupImpl servicegroup.ServiceGroup
}

// CreateRuntimeHPARules create HPA rules, and apply them
func (s *podscalerService) CreateRuntimeHPARules(ctx context.Context, req *pb.HPARuleCreateRequest) (*pb.CommonResponse, error) {
	var (
		userID user.ID
		err    error
	)

	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeHPARules] set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Services) == 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeHPARules] not set rules for any services"))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get userID failed, error: %v", err))
	}

	runtime, err := s.db.GetRuntime(req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get runtime failed, error: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.OperateAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] check permission failed, error: %v", err))
	}

	userInfo, err := s.getUserInfo(userID.String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get user detail info failed, error: %v", err))
	}

	appInfo, err := s.getAppInfo(runtime.ApplicationID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get app detail info failed, error: %v", err))
	}

	if len(req.Services) <= 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeHPARules] failed: not set service"))
	} else {
		if req.Services[0].Deployments == nil || req.Services[0].Deployments.Replicas == 0 {
			uniqueId := spec.RuntimeUniqueId{
				ApplicationId: runtime.ApplicationID,
				Workspace:     runtime.Workspace,
				Name:          runtime.Name,
			}
			preDeploy, err := s.db.GetPreDeployment(uniqueId)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get PreDeployment failed: %v", err))
			}

			var diceObj diceyml.Object
			if preDeploy.DiceOverlay != "" {
				if err = json.Unmarshal([]byte(preDeploy.DiceOverlay), &diceObj); err != nil {
					return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] Unmarshall preDeploy.DiceOverlay failed: %v", err))
				}
			} else {
				if err = json.Unmarshal([]byte(preDeploy.Dice), &diceObj); err != nil {
					return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] Unmarshall preDeploy.Dice failed: %v", err))
				}
			}
			for idx, svc := range req.Services {
				if _, ok := diceObj.Services[svc.ServiceName]; ok {
					req.Services[idx].Deployments = &pb.Deployments{
						Replicas: uint64(diceObj.Services[svc.ServiceName].Deployments.Replicas),
					}
					req.Services[idx].Resources = &pb.Resources{
						Cpu:  diceObj.Services[svc.ServiceName].Resources.CPU,
						Mem:  int64(diceObj.Services[svc.ServiceName].Resources.Mem),
						Disk: 0,
					}
				} else {
					return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] error: service %s not found in PreDeployment", svc.ServiceName))
				}
			}
		}
	}

	err = validateHPARuleConfig(req.Services)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] validate rules failed, error: %v", err))
	}

	return s.createHPARule(userInfo, appInfo, runtime, req.Services)
}

// ListRuntimeHPARules list HPA rules for services in runtime, if no services in req, then list all HPA rules in the runtime
func (s *podscalerService) ListRuntimeHPARules(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeHPARules, error) {
	var (
		userID user.ID
		err    error
	)
	logrus.Infof("grt runtime ID %s hpa rules for services = %s", req.RuntimeId, req.Services)
	if req.RuntimeId == "" {
		return nil, errors.New(fmt.Sprint("[ListRuntimeHPARules] runtimeId not set"))
	}
	reqServices := strings.Split(req.Services, ",")
	//reqServices maybe length as 1 and with empty value
	services := make([]string, 0)

	for _, svc := range reqServices {
		if svc != "" {
			services = append(services, svc)
		}
	}

	runtimeID, err := strconv.ParseUint(req.RuntimeId, 10, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPARules] parse runtimeID failed: %v", err))
	}

	if runtimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ListRuntimeHPARules] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPARules] get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPARules] getruntime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.GetAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPARules] check permission failed: %v", err))
	}

	return s.listHPARules(runtime, services)
}

// UpdateRuntimeHPARules update HPA rules with the target ruleIDs
func (s *podscalerService) UpdateRuntimeHPARules(ctx context.Context, req *pb.ErdaRuntimeHPARules) (*pb.CommonResponse, error) {
	var (
		userID user.ID
		err    error
	)

	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeHPARules] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Rules) == 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeHPARules] no rules set for update"))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] get runtime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.OperateAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] check permission failed: %v", err))
	}

	userInfo, err := s.getUserInfo(userID.String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] get user info failed: %v", err))
	}

	appInfo, err := s.getAppInfo(runtime.ApplicationID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] get app info failed: %v", err))
	}

	// map[id]dbclient.RuntimeHPA
	oldRules := make(map[string]dbclient.RuntimeHPA)
	// map[id]*pb.ScaledConfig
	newRules := make(map[string]*pb.ScaledConfig)
	for _, rule := range req.Rules {
		if rule.ScaledConfig == nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] update hpa rule failed: scaledConfig not set for rule id: %s", rule.RuleID))
		}

		ruleHPA, err := s.db.GetRuntimeHPARuleByRuleId(rule.RuleID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] update hpa rule failed: get rule by rule id %s with error: %v", rule.RuleID, err))
		}

		oldRules[ruleHPA.ID] = ruleHPA
		newRule := &pb.ScaledConfig{}
		err = json.Unmarshal([]byte(ruleHPA.Rules), newRule)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] update hpa rule failed: Unmarshal rule by rule id %s with error: %v", rule.RuleID, err))
		}
		newRules[ruleHPA.ID] = newRule

		err = validateHPARuleConfigCustom(rule.RuleName, 2*newRule.MaxReplicaCount, rule.ScaledConfig)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] update hpa rule failed for svc %s:validate rule by rule id %s with error: %v", ruleHPA.ServiceName, rule.RuleID, err))
		}
	}

	return s.updateHPARules(userInfo, appInfo, runtime, newRules, oldRules, req)
}

// DeleteHPARulesByIds delete HPA rules by target ruleIDs
func (s *podscalerService) DeleteHPARulesByIds(ctx context.Context, req *pb.DeleteRuntimePARulesRequest) (*pb.CommonResponse, error) {
	var (
		userID user.ID
		err    error
	)

	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[DeleteHPARulesByIds] set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[DeleteHPARulesByIds] get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[DeleteHPARulesByIds] get runtime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.OperateAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[DeleteHPARulesByIds] check permission failed: %v", err))
	}

	return s.deleteHPARule(userID.String(), runtime, req.Rules)
}

// ApplyOrCancelHPARulesByIds apply or cancel HPA rules by target ruleIDs
func (s *podscalerService) ApplyOrCancelHPARulesByIds(ctx context.Context, req *pb.ApplyOrCancelPARulesRequest) (*pb.CommonResponse, error) {
	var (
		userID user.ID
		err    error
	)

	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelHPARulesByIds] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.RuleAction) == 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelHPARulesByIds] actions not set in request"))
	}
	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[ApplyOrCancelHPARulesByIds] get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ApplyOrCancelHPARulesByIds] get runtime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.OperateAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ApplyOrCancelHPARulesByIds] check permission failed: %v", err))
	}

	userInfo, err := s.getUserInfo(userID.String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ApplyOrCancelHPARulesByIds] get user info failed: %v", err))
	}
	return s.applyOrCancelHPARule(userInfo, runtime, req.RuleAction)
}

func (s *podscalerService) GetRuntimeBaseInfo(ctx context.Context, req *pb.ListRequest) (*pb.RuntimeServiceBaseInfos, error) {
	var (
		err    error
		userID user.ID
	)
	logrus.Infof("[GetRuntimeBaseInfo] get runtime ID %s hpa rules for services = %s", req.RuntimeId, req.Services)

	if req.RuntimeId == "" {
		return nil, errors.New(fmt.Sprint("[GetRuntimeBaseInfo] runtimeId not set"))
	}

	runtimeID, err := strconv.ParseUint(req.RuntimeId, 10, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] runtimeId %s is not valid id, runtimeId must bigger than 0", req.RuntimeId))
	}

	if runtimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[GetRuntimeBaseInfo] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	runtime, err := s.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] get runtime failed: %v", err))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] get userID and OrgID failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.GetAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] check permission failed: %v", err))
	}

	id := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}

	preDeploy, err := s.db.GetPreDeployment(id)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] get DetPreDeployment failed: %v", err))
	}

	var diceObj diceyml.Object
	if err = json.Unmarshal([]byte(preDeploy.Dice), &diceObj); err != nil {
		return nil, errors.New(fmt.Sprintf("[GetRuntimeBaseInfo] Unmarshall dice failed: %v", err))
	}

	svcInfos := make([]*pb.ServiceBaseInfo, 0)
	for name, svc := range diceObj.Services {
		svcInfos = append(svcInfos, &pb.ServiceBaseInfo{
			ServiceName: name,
			Deployments: &pb.Deployments{
				Replicas: uint64(svc.Deployments.Replicas),
			},
			Resources: &pb.Resources{
				Cpu: svc.Resources.CPU,
				Mem: int64(svc.Resources.Mem),
			},
		})
	}
	return &pb.RuntimeServiceBaseInfos{
		RuntimeID:        runtimeID,
		ServiceBaseInfos: svcInfos,
	}, nil
}

func (s *podscalerService) ListRuntimeHPAEvents(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeHPAEvents, error) {
	var (
		userID user.ID
		err    error
	)
	logrus.Infof("get runtime ID %s hpa rules for services = %s", req.RuntimeId, req.Services)
	if req.RuntimeId == "" {
		return nil, errors.New(fmt.Sprint("[ListRuntimeHPAEvents] runtimeId not set"))
	}
	reqServices := strings.Split(req.Services, ",")
	//reqServices maybe length as 1 and with empty value
	services := make([]string, 0)

	for _, svc := range reqServices {
		if svc != "" {
			services = append(services, svc)
		}
	}

	runtimeID, err := strconv.ParseUint(req.RuntimeId, 10, 64)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPAEvents] parse runtimeID failed: %v", err))
	}

	if runtimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ListRuntimeHPAEvents] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPAEvents] get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPAEvents] getruntime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.GetAction)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPAEvents] check permission failed: %v", err))
	}

	return s.listHPAEvents(runtime.ID, services)
}

func (s *podscalerService) CreateRuntimeVPARules(ctx context.Context, req *pb.VPARuleCreateRequest) (*pb.CommonResponse, error) {
	return nil, nil
}
func (s *podscalerService) ListRuntimeVPARules(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeVPARules, error) {
	return nil, nil
}
func (s *podscalerService) UpdateRuntimeVPARules(ctx context.Context, req *pb.ErdaRuntimeVPARules) (*pb.CommonResponse, error) {
	return nil, nil
}
func (s *podscalerService) DeleteVPARulesByIds(ctx context.Context, req *pb.DeleteRuntimePARulesRequest) (*pb.CommonResponse, error) {
	return nil, nil
}
func (s *podscalerService) ApplyOrCancelVPARulesByIds(ctx context.Context, req *pb.ApplyOrCancelPARulesRequest) (*pb.CommonResponse, error) {
	return nil, nil
}
func (s *podscalerService) ListRuntimeVPARecommendations(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeVPARecommendations, error) {
	return nil, nil
}

func (s *podscalerService) HPScaleManual(ctx context.Context, req *pb.ManualHPRequest) (*pb.HPManualResponse, error) {
	appId_ := req.ApplicationId
	appId, err := strconv.Atoi(appId_)
	if err != nil {
		return nil, errors.Errorf("[HPScaleManual] failed to update Overlay, appId invalid: %v", appId_)
	}

	workspace := req.Workspace
	if workspace == "" {
		return nil, errors.Errorf("[HPScaleManual] workspace not set")
	}

	userID, _, err := s.GetUserAndOrgID(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[HPScaleManual] get userID and orgID failed: %v", err))
	}

	perm, err := s.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(appId),
		Resource: "runtime-" + strutil.ToLower(workspace),
		Action:   apistructs.OperateAction,
	})

	if err != nil {
		return nil, errors.New(fmt.Sprintf("[HPScaleManual] check application permission for userID %s error: %v", userID.String(), err))
	}

	if !perm.Access {
		return nil, errors.New(fmt.Sprintf("[HPScaleManual] no permission for userID %s", userID.String()))
	}

	runtimeName := req.RuntimeName
	if runtimeName == "" {
		return nil, errors.Errorf("[HPScaleManual] runtimeName invalid: not set runtimeName")
	}

	rsr := pb.RuntimeScaleRecord{
		ApplicationId: uint64(appId),
		Workspace:     workspace,
		Name:          runtimeName,
		Payload: &pb.PreDiceDTO{
			Name:     req.Name,
			Envs:     req.Envs,
			Services: req.Services,
		},
	}

	oldOverlayDataForAudit, err := s.processRuntimeScaleRecord(rsr, "")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[HPScaleManual] process runtime scale record failed: %v", err))
	}

	return &pb.HPManualResponse{
		Name:     oldOverlayDataForAudit.Name,
		Envs:     oldOverlayDataForAudit.Envs,
		Services: oldOverlayDataForAudit.Services,
	}, nil
}

func (s *podscalerService) BatchHPScaleManual(ctx context.Context, req *pb.BatchManualHPRequest) (*pb.BatchManualResponse, error) {
	// TODO .
	return nil, status.Errorf(codes.Unimplemented, "method BatchHPScaleManual not implemented")
}

type ServiceOption func(*podscalerService) *podscalerService

func WithBundleService(s BundleService) ServiceOption {
	return func(service *podscalerService) *podscalerService {
		service.bundle = s
		return service
	}
}

func WithDBService(db DBService) ServiceOption {
	return func(service *podscalerService) *podscalerService {
		service.db = db
		return service
	}
}

func WithServiceGroupImpl(serviceGroupImpl servicegroup.ServiceGroup) ServiceOption {
	return func(service *podscalerService) *podscalerService {
		service.serviceGroupImpl = serviceGroupImpl
		return service
	}
}

func NewRuntimeHPScalerService(options ...ServiceOption) pb.PodScalerServiceServer {
	s := &podscalerService{}

	for _, option := range options {
		option(s)
	}

	return s
}
