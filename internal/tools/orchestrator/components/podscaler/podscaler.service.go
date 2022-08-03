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
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeHPARules] set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Services) == 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeHPARules] not set rules for any services"))
	}

	runtime, userInfo, appInfo, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] %v", err))
	}

	if req.Services[0].Deployments == nil || req.Services[0].Deployments.Replicas == 0 || req.Services[0].Resources == nil {
		// not set resources for service, get from PreDeployment
		err = s.initReplicasAndResources(runtime, req, nil, true)
		if err != nil {
			return nil, errors.Errorf("[CreateRuntimeHPARules] init replicas and resources error: %v", err)
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
	logrus.Infof("[ListRuntimeHPARules] get runtime ID %s hpa rules for services = %s", req.RuntimeId, req.Services)
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

	runtime, err := s.checkPermission(ctx, runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPARules] check permission failed: %v", err))
	}

	return s.listHPARules(runtime, services)
}

// UpdateRuntimeHPARules update HPA rules with the target ruleIDs
func (s *podscalerService) UpdateRuntimeHPARules(ctx context.Context, req *pb.ErdaRuntimeHPARules) (*pb.CommonResponse, error) {
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeHPARules] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Rules) == 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeHPARules] no rules set for update"))
	}

	runtime, userInfo, appInfo, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeHPARules] %v", err))
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

	runtime, err := s.checkPermission(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[DeleteHPARulesByIds] check permission failed: %v", err))
	}

	return s.deleteHPARule(userID.String(), runtime, req.Rules)
}

// ApplyOrCancelHPARulesByIds apply or cancel HPA rules by target ruleIDs
func (s *podscalerService) ApplyOrCancelHPARulesByIds(ctx context.Context, req *pb.ApplyOrCancelPARulesRequest) (*pb.CommonResponse, error) {
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelHPARulesByIds] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.RuleAction) == 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelHPARulesByIds] actions not set in request"))
	}

	runtime, userInfo, _, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ApplyOrCancelHPARulesByIds] %v", err))
	}

	return s.applyOrCancelHPARule(userInfo, runtime, req.RuleAction)
}

func (s *podscalerService) GetRuntimeBaseInfo(ctx context.Context, req *pb.ListRequest) (*pb.RuntimeServiceBaseInfos, error) {
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

	runtime, err := s.checkPermission(ctx, runtimeID)
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

	runtime, err := s.checkPermission(ctx, runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeHPAEvents] check permission failed: %v", err))
	}

	return s.listHPAEvents(runtime.ID, services)
}

func (s *podscalerService) CreateRuntimeVPARules(ctx context.Context, req *pb.VPARuleCreateRequest) (*pb.CommonResponse, error) {
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeVPARules] set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Services) == 0 {
		return nil, errors.New(fmt.Sprint("[CreateRuntimeVPARules] not set rules for any services"))
	}

	runtime, userInfo, appInfo, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeVPARules] %v", err))
	}

	if req.Services[0].Deployments == nil || req.Services[0].Deployments.Replicas == 0 || req.Services[0].Resources == nil {
		// not set resources for service, get from PreDeployment
		err = s.initReplicasAndResources(runtime, nil, req, false)
		if err != nil {
			return nil, errors.Errorf("[CreateRuntimeHPARules] init replicas and resources error: %v", err)
		}
	}

	err = validateVPARuleCreateConfig(req.Services, false)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[CreateRuntimeVPARules] validate rules failed, error: %v", err))
	}

	return s.createVPARule(userInfo, appInfo, runtime, req.Services)

}

func (s *podscalerService) ListRuntimeVPARules(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeVPARules, error) {
	logrus.Infof("[ListRuntimeVPARules] get runtime ID %s vpa rules for services = %s", req.RuntimeId, req.Services)
	if req.RuntimeId == "" {
		return nil, errors.New(fmt.Sprint("[ListRuntimeVPARules] runtimeId not set"))
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
		return nil, errors.New(fmt.Sprintf("[ListRuntimeVPARules] parse runtimeID failed: %v", err))
	}

	if runtimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ListRuntimeVPARules] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	runtime, err := s.checkPermission(ctx, runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeVPARules] check permission failed: %v", err))
	}

	return s.listVPARules(runtime, services)
}

func (s *podscalerService) UpdateRuntimeVPARules(ctx context.Context, req *pb.ErdaRuntimeVPARules) (*pb.CommonResponse, error) {
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeVPARules] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.Rules) == 0 {
		return nil, errors.New(fmt.Sprint("[UpdateRuntimeVPARules] no rules set for update"))
	}

	runtime, userInfo, appInfo, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeVPARules] %v", err))
	}

	// map[id]dbclient.RuntimeHPA
	oldVPAs := make(map[string]dbclient.RuntimeVPA)
	oldRules := make(map[string]*pb.RuntimeServiceVPAConfig)
	// map[id]*pb.ScaledConfig
	newRules := make(map[string]*pb.RuntimeServiceVPAConfig)
	updateRules := make([]*pb.RuntimeServiceVPAConfig, 0)
	for _, rule := range req.Rules {
		if rule.Rule == nil {
			return nil, errors.Errorf("[UpdateRuntimeVPARules] update vpa rule failed: rule not set")
		}

		ruleVPA, err := s.db.GetRuntimeVPARuleByRuleId(rule.Rule.RuleID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeVPARules] update vpa rule failed: get rule by rule id %s with error: %v", rule.Rule.RuleID, err))
		}

		oldVPAs[ruleVPA.ID] = ruleVPA
		oldRule := &pb.RuntimeServiceVPAConfig{}
		err = json.Unmarshal([]byte(ruleVPA.Rules), oldRule)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("[UpdateRuntimeVPARules] update vpa rule failed: Unmarshal rule by rule id %s with error: %v", rule.Rule.RuleID, err))
		}
		oldRules[ruleVPA.ID] = oldRule
		newRules[ruleVPA.ID] = rule.Rule

		updateRules = append(updateRules, rule.Rule)
	}

	err = validateVPARuleCreateConfig(updateRules, true)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[UpdateRuntimeVPARules] validate rules failed, error: %v", err))
	}

	return s.updateVPARules(userInfo, appInfo, runtime, newRules, oldRules, oldVPAs, nil)
}

func (s *podscalerService) DeleteVPARulesByIds(ctx context.Context, req *pb.DeleteRuntimePARulesRequest) (*pb.CommonResponse, error) {

	var (
		userID user.ID
		err    error
	)

	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[DeleteVPARulesByIds] set invalid runtimeId, runtimeId must bigger than 0"))
	}

	runtime, err := s.checkPermission(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[DeleteVPARulesByIds] check permission failed: %v", err))
	}

	return s.deleteVPARule(userID.String(), runtime, req.Rules)
}

func (s *podscalerService) ApplyOrCancelVPARulesByIds(ctx context.Context, req *pb.ApplyOrCancelPARulesRequest) (*pb.CommonResponse, error) {
	if req.RuntimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelVPARulesByIds] runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	if len(req.RuleAction) == 0 {
		return nil, errors.New(fmt.Sprint("[ApplyOrCancelVPARulesByIds] actions not set in request"))
	}

	runtime, userInfo, _, err := s.getRuntimeDetails(ctx, req.RuntimeID)
	if err != nil {
		return nil, errors.Errorf("[ApplyOrCancelVPARulesByIds] error: %v", err)
	}

	return s.applyOrCancelVPARule(userInfo, runtime, req.RuleAction)
}

func (s *podscalerService) ListRuntimeVPARecommendations(ctx context.Context, req *pb.ListRequest) (*pb.ErdaRuntimeVPARecommendations, error) {
	if req.RuntimeId == "" {
		return nil, errors.New(fmt.Sprint("[ListRuntimeVPARecommendations] runtimeId not set"))
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
		return nil, errors.New(fmt.Sprintf("[ListRuntimeVPARecommendations] parse runtimeID failed: %v", err))
	}

	if runtimeID <= 0 {
		return nil, errors.New(fmt.Sprint("[ListRuntimeVPARecommendations]runtime not set or set invalid runtimeId, runtimeId must bigger than 0"))
	}

	runtime, err := s.checkPermission(ctx, runtimeID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[ListRuntimeVPARecommendations] check permission failed: %v", err))
	}

	return s.listVPAServiceRecommendations(runtime.ID, services)
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
