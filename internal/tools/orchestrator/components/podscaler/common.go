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

	"github.com/pkg/errors"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	pstypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *podscalerService) GetUserAndOrgID(ctx context.Context) (userID user.ID, orgID uint64, err error) {
	orgIntID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		err = apierrors.ErrGetRuntime.InvalidParameter(errors.New("Org-ID"))
		return
	}

	orgID = uint64(orgIntID)

	userID = user.ID(apis.GetUserID(ctx))
	if userID.Invalid() {
		err = apierrors.ErrGetRuntime.NotLogin()
		return
	}

	return
}

func (s *podscalerService) checkRuntimeScopePermission(userID user.ID, runtime *dbclient.Runtime, action string) error {
	perm, err := s.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   action,
	})

	if err != nil {
		return err
	}

	if !perm.Access {
		return apierrors.ErrGetRuntime.AccessDenied()
	}

	return nil
}

func (s *podscalerService) getUserInfo(userID string) (*apistructs.UserInfo, error) {
	return s.bundle.GetCurrentUser(userID)
}

func (s *podscalerService) getAppInfo(id uint64) (*apistructs.ApplicationDTO, error) {
	return s.bundle.GetApp(id)
}

func (s *podscalerService) getRuntimeDetails(ctx context.Context, runtimeId uint64) (*dbclient.Runtime, *apistructs.UserInfo, *apistructs.ApplicationDTO, error) {
	var (
		userID user.ID
		err    error
	)

	if userID, _, err = s.GetUserAndOrgID(ctx); err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("get userID failed, error: %v", err))
	}

	runtime, err := s.db.GetRuntime(runtimeId)
	if err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("get runtime failed, error: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.OperateAction)
	if err != nil {
		return runtime, nil, nil, errors.New(fmt.Sprintf("check permission failed, error: %v", err))
	}

	userInfo, err := s.getUserInfo(userID.String())
	if err != nil {
		return runtime, nil, nil, errors.New(fmt.Sprintf("get user detail info failed, error: %v", err))
	}

	appInfo, err := s.getAppInfo(runtime.ApplicationID)
	if err != nil {
		return runtime, userInfo, nil, errors.New(fmt.Sprintf("[CreateRuntimeHPARules] get app detail info failed, error: %v", err))
	}

	return runtime, userInfo, appInfo, nil
}

func (s *podscalerService) checkPermission(ctx context.Context, runtimeId uint64) (*dbclient.Runtime, error) {
	userID, _, err := s.GetUserAndOrgID(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("get userID and orgID failed: %v", err))
	}

	runtime, err := s.db.GetRuntime(runtimeId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("getruntime failed: %v", err))
	}

	err = s.checkRuntimeScopePermission(userID, runtime, apistructs.GetAction)
	if err != nil {
		return runtime, errors.New(fmt.Sprintf("checkRuntimeScopePermission failed: %v", err))
	}
	return runtime, nil
}

func (s *podscalerService) initReplicasAndResources(runtime *dbclient.Runtime, req *pb.HPARuleCreateRequest, reqVPA *pb.VPARuleCreateRequest, isHPA bool) error {
	uniqueId := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}
	preDeploy, err := s.db.GetPreDeployment(uniqueId)
	if err != nil {
		return errors.New(fmt.Sprintf("get PreDeployment failed: %v", err))
	}

	var diceObj diceyml.Object
	if preDeploy.DiceOverlay != "" {
		if err = json.Unmarshal([]byte(preDeploy.DiceOverlay), &diceObj); err != nil {
			return errors.New(fmt.Sprintf("Unmarshall preDeploy.DiceOverlay failed: %v", err))
		}
	} else {
		if err = json.Unmarshal([]byte(preDeploy.Dice), &diceObj); err != nil {
			return errors.New(fmt.Sprintf("Unmarshall preDeploy.Dice failed: %v", err))
		}
	}
	if isHPA {
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
				return errors.New(fmt.Sprintf("error: service %s not found in PreDeployment", svc.ServiceName))
			}
		}
	} else {
		for idx, svc := range reqVPA.Services {
			if _, ok := diceObj.Services[svc.ServiceName]; ok {
				reqVPA.Services[idx].Deployments = &pb.Deployments{
					Replicas: uint64(diceObj.Services[svc.ServiceName].Deployments.Replicas),
				}
				reqVPA.Services[idx].Resources = &pb.Resources{
					Cpu: diceObj.Services[svc.ServiceName].Resources.CPU,
					Mem: int64(diceObj.Services[svc.ServiceName].Resources.Mem),
				}
			} else {
				return errors.New(fmt.Sprintf("error: service %s not found in PreDeployment", svc.ServiceName))
			}
		}
	}
	return nil
}

func (s *podscalerService) applyOrCancelRule(runtime *dbclient.Runtime, hpaRule *dbclient.RuntimeHPA, vpaRule *dbclient.RuntimeVPA, ruleId string, action string, isHPA bool) error {
	id := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
	}

	namespace, name := runtime.ScheduleName.Args()
	sg := &apistructs.ServiceGroup{
		ClusterName: runtime.ClusterName,
		Dice: apistructs.Dice{
			ID:       name,
			Type:     namespace,
			Services: make([]apistructs.Service, 0),
		},
		Extra: make(map[string]string),
	}
	sg.Labels = make(map[string]string)
	sg.Labels[pstypes.ErdaPALabelKey] = action

	if isHPA {
		if hpaRule == nil {
			rule, err := s.db.GetRuntimeHPARuleByRuleId(ruleId)
			if err != nil {
				return err
			}
			hpaRule = &rule
		}
		if action != pstypes.ErdaHPALabelValueCancel {
			serviceVPAs, err := s.db.GetRuntimeVPARulesByServices(id, []string{hpaRule.ServiceName})
			if err != nil {
				return err
			}
			for _, serviceVpaRule := range serviceVPAs {
				if serviceVpaRule.IsApplied == pstypes.RuntimePARuleApplied {
					scaledConfig := pb.ScaledConfig{}
					err = json.Unmarshal([]byte(hpaRule.Rules), &scaledConfig)
					if err != nil {
						return err
					}
					for _, trigger := range scaledConfig.Triggers {
						if trigger.Type == pstypes.ErdaHPATriggerCPU || trigger.Type == pstypes.ErdaHPATriggerMemory {
							return errors.Errorf("Service %s hpa has cpu and/or memory trigger, can not apply when vpa enable", hpaRule.ServiceName)
						}
					}
				}
			}
		}

		sg.Services = append(sg.Services, apistructs.Service{
			Name: hpaRule.ServiceName,
		})
		sg.Extra[hpaRule.ServiceName] = hpaRule.Rules
	} else {
		if vpaRule == nil {
			rule, err := s.db.GetRuntimeVPARuleByRuleId(ruleId)
			if err != nil {
				return err
			}
			vpaRule = &rule
		}
		if action != pstypes.ErdaVPALabelValueCancel {
			serviceHPAs, err := s.db.GetRuntimeHPARulesByServices(id, []string{vpaRule.ServiceName})
			if err != nil {
				return err
			}
			for _, serviceHpaRule := range serviceHPAs {
				if serviceHpaRule.IsApplied == pstypes.RuntimePARuleApplied {
					scaledConfig := pb.ScaledConfig{}
					err = json.Unmarshal([]byte(serviceHpaRule.Rules), &scaledConfig)
					if err != nil {
						return err
					}
					for _, trigger := range scaledConfig.Triggers {
						if trigger.Type == pstypes.ErdaHPATriggerCPU || trigger.Type == pstypes.ErdaHPATriggerMemory {
							return errors.Errorf("Service %s hpa has applied with cpu and/or memory trigger, can not apply vpa rule", vpaRule.ServiceName)
						}
					}
				}
			}
		}
		sg.Services = append(sg.Services, apistructs.Service{
			Name: vpaRule.ServiceName,
		})
		sg.Extra[vpaRule.ServiceName] = vpaRule.Rules
	}

	_, err := s.serviceGroupImpl.Scale(sg)
	if err != nil {
		return err
	}
	return nil
}

func (s *podscalerService) getTargetMeta(runtime *dbclient.Runtime, serviceHPARules []*pb.RuntimeServiceHPAConfig, serviceVPARules []*pb.RuntimeServiceVPAConfig, isHPA bool) (map[string]int, map[string]pstypes.ErdaHPAObject, error) {
	uniqueId := spec.RuntimeUniqueId{
		ApplicationId: runtime.ApplicationID,
		Workspace:     runtime.Workspace,
		Name:          runtime.Name,
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
	sg.Labels = make(map[string]string)
	sg.Labels[pstypes.ErdaPALabelKey] = pstypes.ErdaHPALabelValueCreate

	mapSVCNameToIndx := make(map[string]int)
	if isHPA {
		for idx, hpaRule := range serviceHPARules {
			mapSVCNameToIndx[hpaRule.ServiceName] = idx
			sg.Services = append(sg.Services, apistructs.Service{
				Name:  hpaRule.ServiceName,
				Scale: int(hpaRule.Deployments.Replicas),
				Resources: apistructs.Resources{
					Cpu:  hpaRule.Resources.Cpu,
					Mem:  float64(hpaRule.Resources.Mem),
					Disk: float64(hpaRule.Resources.Disk),
				},
			})
		}
	} else {
		for idx, vpaRule := range serviceVPARules {
			mapSVCNameToIndx[vpaRule.ServiceName] = idx
			sg.Services = append(sg.Services, apistructs.Service{
				Name: vpaRule.ServiceName,
				Resources: apistructs.Resources{
					Cpu:  vpaRule.Resources.Cpu,
					Mem:  float64(vpaRule.Resources.Mem),
					Disk: float64(vpaRule.Resources.Disk),
				},
			})
		}
	}

	sgHPAObjects, err := s.serviceGroupImpl.Scale(sg)
	if err != nil {
		return nil, nil, errors.Errorf("get targetRef for service %s for runtime %s for runtime %#v failed for servicegroup, err: %v", sg.Services[0].Name, uniqueId.Name, uniqueId, err)
	}

	sgSvcObject, ok := sgHPAObjects.(map[string]pstypes.ErdaHPAObject)
	if !ok {
		return nil, nil, errors.Errorf("get targetRef Meta for service %s for runtime %s for runtimeUniqueId %#v failed: return is not an map[string]pstypes.ErdaHPAObject object", sg.Services[0].Name, uniqueId.Name, uniqueId)
	}
	return mapSVCNameToIndx, sgSvcObject, nil
}
