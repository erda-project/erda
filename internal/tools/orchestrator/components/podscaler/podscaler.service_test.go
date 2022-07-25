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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/user"
	patypes "github.com/erda-project/erda/internal/tools/orchestrator/components/podscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

func Test_podScalerService_CreateRuntimeHPARules(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}

	type args struct {
		ctx context.Context
		req *pb.HPARuleCreateRequest
	}

	tTime := time.Now()
	services := make([]*pb.RuntimeServiceHPAConfig, 0)
	metadata := make(map[string]string)
	metadata["type"] = "Utilization"
	metadata["value"] = "20"
	triggers := make([]*pb.ScaleTriggers, 0)
	triggers = append(triggers, &pb.ScaleTriggers{
		Type:     "memory",
		Metadata: metadata,
	})

	services01 := make([]*pb.RuntimeServiceHPAConfig, 0)
	metadata01 := make(map[string]string)
	metadata01["type"] = "Utilization"
	metadata01["value"] = "200"
	triggers01 := make([]*pb.ScaleTriggers, 0)
	triggers01 = append(triggers01, &pb.ScaleTriggers{
		Type:     "memory",
		Metadata: metadata01,
	})

	services02 := make([]*pb.RuntimeServiceHPAConfig, 0)
	metadata02 := make(map[string]string)
	metadata02["timezone"] = "Asia/Shanghai"
	metadata02["start"] = "10 * * * *"
	metadata02["end"] = "50 * * * *"
	metadata02["desiredReplicas"] = "10"
	triggers02 := make([]*pb.ScaleTriggers, 0)
	triggers02 = append(triggers02, &pb.ScaleTriggers{
		Type:     "cron",
		Metadata: metadata02,
	})

	services = append(services, &pb.RuntimeServiceHPAConfig{
		RuleName: "test01",
		Deployments: &pb.Deployments{
			Replicas: 1,
		},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
		ScaledConfig: &pb.ScaledConfig{
			MaxReplicaCount: 3,
			MinReplicaCount: 1,
			Advanced: &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			},
			Triggers: triggers,
		},
	})
	services01 = append(services01, &pb.RuntimeServiceHPAConfig{
		RuleName: "test01",
		Deployments: &pb.Deployments{
			Replicas: 1,
		},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
		ScaledConfig: &pb.ScaledConfig{
			MaxReplicaCount: 3,
			MinReplicaCount: 1,
			Advanced: &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			},
			Triggers: triggers01,
		},
	})
	services02 = append(services02, &pb.RuntimeServiceHPAConfig{
		RuleName: "test01",
		Deployments: &pb.Deployments{
			Replicas: 1,
		},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
		ScaledConfig: &pb.ScaledConfig{
			MaxReplicaCount: 3,
			MinReplicaCount: 1,
			Advanced: &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			},
			Triggers: triggers02,
		},
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CommonResponse
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.HPARuleCreateRequest{
					RuntimeID: 1,
					Services:  services,
					// TODO: setup fields
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.HPARuleCreateRequest{
					RuntimeID: 1,
					Services:  services01,
					// TODO: setup fields
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test_03",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.HPARuleCreateRequest{
					RuntimeID: 1,
					Services:  services02,
					// TODO: setup fields
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "GetCurrentUser",
				func(_ *bundle.Bundle, userID string) (*apistructs.UserInfo, error) {
					return &apistructs.UserInfo{
						ID:   "1",
						Name: "name",
						Nick: "nick",
					}, nil
				})

			m5 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "GetApp",
				func(_ *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
					return &apistructs.ApplicationDTO{
						ID:                 1,
						Name:               "test",
						DisplayName:        "test",
						OrgID:              1,
						OrgName:            "test",
						OrgDisplayName:     "test",
						ProjectID:          1,
						ProjectName:        "test",
						ProjectDisplayName: "test",
					}, nil
				})

			m6 := monkey.PatchInstanceMethod(reflect.TypeOf(s.serviceGroupImpl), "Scale",
				func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {

					ret := make(map[string]patypes.ErdaHPAObject)
					ret["test01"] = patypes.ErdaHPAObject{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Deployment",
							APIVersion: "apps/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-9ddcc8b369",
							Namespace: "project-1-prod",
						},
					}
					return ret, nil
				})

			m7 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "CreateHPARule",
				func(_ *dbServiceImpl, req *dbclient.RuntimeHPA) error {
					return nil
				})

			m8 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARuleByRuleId",
				func(_ *dbServiceImpl, ruleId string) (dbclient.RuntimeHPA, error) {
					return generateRuntimeHPA(tTime), nil
				})

			defer m8.Unpatch()
			defer m7.Unpatch()
			defer m6.Unpatch()
			defer m5.Unpatch()
			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.CreateRuntimeHPARules(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRuntimeHPARules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateRuntimeHPARules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podScalerService_ListRuntimeHPARules(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.ListRequest
	}

	metadata := make(map[string]string)
	metadata["type"] = "Utilization"
	metadata["value"] = "20"
	triggers := make([]*pb.ScaleTriggers, 0)
	triggers = append(triggers, &pb.ScaleTriggers{
		Type:     "memory",
		Metadata: metadata,
	})

	tTime := time.Now()
	rules := make([]*pb.ErdaRuntimeHPARule, 0)
	rules = append(rules, &pb.ErdaRuntimeHPARule{
		RuleID:      "1779dff5-184c-4dfe-9c76-978ac5126e59",
		CreateAt:    timestamppb.New(tTime),
		UpdateAt:    timestamppb.New(tTime),
		ServiceName: "service01",
		RuleName:    "test",
		UserInfo: &pb.UserInfo{
			UserID:       1,
			UserName:     "name",
			UserNickName: "nick",
		},
		ScaledConfig: &pb.ScaledConfig{
			RuleName:      "test",
			RuleNameSpace: "project-3-prod",
			ScaleTargetRef: &pb.ScaleTargetRef{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-cdd1a9f7c4",
			},
			MinReplicaCount: 1,
			MaxReplicaCount: 3,
			Advanced: &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			},
			Triggers: triggers,
			Fallback: &pb.FallBack{
				Replicas: 1,
			},
		},
		IsApplied: "Y",
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ErdaRuntimeHPARules
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ListRequest{
					RuntimeId: "1",
				},
			},
			want: &pb.ErdaRuntimeHPARules{
				RuntimeID: 1,
				Rules:     rules,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})
			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARulesByServices",
				func(_ *dbServiceImpl, id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeHPA, error) {
					rules := make([]dbclient.RuntimeHPA, 0)
					rules = append(rules, generateRuntimeHPA(tTime))
					return rules, nil
				})

			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.ListRuntimeHPARules(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRuntimeHPARules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			str01, _ := json.Marshal(got)
			str02, _ := json.Marshal(tt.want)
			fmt.Printf("str01:%s\n", string(str01))
			fmt.Printf("str02:%s\n", string(str02))
			if string(str01) != string(str02) {
				t.Errorf("ListRuntimeHPARules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podScalerService_DeleteHPARulesByIds(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteRuntimePARulesRequest
	}

	tTime := time.Now()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CommonResponse
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.DeleteRuntimePARulesRequest{
					RuntimeID: 1,
					//Rules:       nil,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARulesByRuntimeId",
				func(_ *dbServiceImpl, runtimeID uint64) ([]dbclient.RuntimeHPA, error) {
					rules := make([]dbclient.RuntimeHPA, 0)
					rules = append(rules, generateRuntimeHPA(tTime))
					return rules, nil
				})

			m5 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARuleByRuleId",
				func(_ *dbServiceImpl, ruleId string) (dbclient.RuntimeHPA, error) {
					return generateRuntimeHPA(tTime), nil
				})
			m6 := monkey.PatchInstanceMethod(reflect.TypeOf(s.serviceGroupImpl), "Scale",
				func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {
					return nil, nil
				})

			m7 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "DeleteRuntimeHPARulesByRuleId",
				func(_ *dbServiceImpl, ruleId string) error {
					return nil
				})
			m8 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "DeleteRuntimeHPAEventsByRuleId",
				func(_ *dbServiceImpl, ruleId string) error {
					return nil
				})

			defer m8.Unpatch()
			defer m7.Unpatch()
			defer m6.Unpatch()
			defer m5.Unpatch()
			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.DeleteHPARulesByIds(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteHPARulesByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteHPARulesByIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podScalerService_UpdateRuntimeHPARules(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.ErdaRuntimeHPARules
	}

	tTime := time.Now()
	metadata := make(map[string]string)
	metadata["type"] = "Utilization"
	metadata["value"] = "23"
	triggers := make([]*pb.ScaleTriggers, 0)
	triggers = append(triggers, &pb.ScaleTriggers{
		Type:     "memory",
		Metadata: metadata,
	})
	updateRules := make([]*pb.ErdaRuntimeHPARule, 0)
	updateRules = append(updateRules, &pb.ErdaRuntimeHPARule{
		RuleID:   "1779dff5-184c-4dfe-9c76-978ac5126e59",
		CreateAt: timestamppb.New(tTime),
		UpdateAt: timestamppb.New(tTime),
		UserInfo: &pb.UserInfo{
			UserID: 1,
		},
		ScaledConfig: &pb.ScaledConfig{
			RuleName:       "",
			RuleNameSpace:  "",
			ScaleTargetRef: nil,

			MinReplicaCount: 1,
			MaxReplicaCount: 6,
			Advanced: &pb.HPAAdvanced{
				RestoreToOriginalReplicaCount: true,
			},
			Triggers: triggers,
			Fallback: &pb.FallBack{
				Replicas: 1,
			},
		},
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CommonResponse
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ErdaRuntimeHPARules{
					RuntimeID: 1,
					Rules:     updateRules,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "GetCurrentUser",
				func(_ *bundle.Bundle, userID string) (*apistructs.UserInfo, error) {
					return &apistructs.UserInfo{
						ID:   "1",
						Name: "name",
						Nick: "nick",
					}, nil
				})
			m5 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARuleByRuleId",
				func(_ *dbServiceImpl, ruleId string) (dbclient.RuntimeHPA, error) {
					return generateRuntimeHPA(tTime), nil
				})

			m6 := monkey.PatchInstanceMethod(reflect.TypeOf(s.serviceGroupImpl), "Scale",
				func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {
					return nil, nil
				})

			m7 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "UpdateHPARule",
				func(_ *dbServiceImpl, req *dbclient.RuntimeHPA) error {
					return nil
				})
			m8 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "GetApp",
				func(_ *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
					return &apistructs.ApplicationDTO{
						ID:                 1,
						Name:               "test",
						DisplayName:        "test",
						OrgID:              1,
						OrgName:            "test",
						OrgDisplayName:     "test",
						ProjectID:          1,
						ProjectName:        "test",
						ProjectDisplayName: "test",
					}, nil
				})
			defer m8.Unpatch()
			defer m7.Unpatch()
			defer m6.Unpatch()
			defer m5.Unpatch()
			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.UpdateRuntimeHPARules(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRuntimeHPARules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRuntimeHPARules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podScalerService_ApplyOrCancelHPARulesByIds(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.ApplyOrCancelPARulesRequest
	}

	tTime := time.Now()
	actions1 := make([]*pb.RuleAction, 0)
	actions2 := make([]*pb.RuleAction, 0)
	actions1 = append(actions1, &pb.RuleAction{
		RuleId: "1779dff5-184c-4dfe-9c76-978ac5126e59",
		Action: "apply",
	})
	actions2 = append(actions2, &pb.RuleAction{
		RuleId: "1779dff5-184c-4dfe-9c76-978ac5126e59",
		Action: "cancel",
	})
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.CommonResponse
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ApplyOrCancelPARulesRequest{
					RuntimeID:  1,
					RuleAction: actions1,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ApplyOrCancelPARulesRequest{
					RuntimeID:  1,
					RuleAction: actions2,
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "GetCurrentUser",
				func(_ *bundle.Bundle, userID string) (*apistructs.UserInfo, error) {
					return &apistructs.UserInfo{
						ID:   "1",
						Name: "name",
						Nick: "nick",
					}, nil
				})

			m5 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARuleByRuleId",
				func(_ *dbServiceImpl, ruleId string) (dbclient.RuntimeHPA, error) {
					rule := generateRuntimeHPA(tTime)
					if tt.name == "Test_01" {
						rule.IsApplied = "N"
					}
					return rule, nil
				})

			m6 := monkey.PatchInstanceMethod(reflect.TypeOf(s.serviceGroupImpl), "Scale",
				func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {
					return nil, nil
				})

			m7 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "UpdateHPARule",
				func(_ *dbServiceImpl, req *dbclient.RuntimeHPA) error {
					return nil
				})

			defer m7.Unpatch()
			defer m6.Unpatch()
			defer m5.Unpatch()
			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.ApplyOrCancelHPARulesByIds(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyOrCancelHPARulesByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyOrCancelHPARulesByIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateRuntimeHPA(tTime time.Time) dbclient.RuntimeHPA {
	return dbclient.RuntimeHPA{
		ID:                     "1779dff5-184c-4dfe-9c76-978ac5126e59",
		CreatedAt:              tTime,
		UpdatedAt:              tTime,
		RuleName:               "test",
		RuleNameSpace:          "project-3-prod",
		OrgID:                  1,
		OrgName:                "test",
		OrgDisPlayName:         "test",
		ProjectID:              1,
		ProjectName:            "test",
		ProjectDisplayName:     "test",
		ApplicationID:          1,
		ApplicationName:        "test",
		ApplicationDisPlayName: "test",
		RuntimeID:              1,
		RuntimeName:            "master",
		ClusterName:            "test",
		Workspace:              "PROD",
		UserID:                 "1",
		UserName:               "name",
		NickName:               "nick",
		ServiceName:            "service01",
		Rules:                  "{\"ruleName\":\"test\",\"ruleNameSpace\":\"project-3-prod\",\"scaleTargetRef\":{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"name\":\"test-cdd1a9f7c4\",\"envSourceContainerName\":\"\"},\"pollingInterval\":0,\"cooldownPeriod\":0,\"minReplicaCount\":1,\"maxReplicaCount\":3,\"advanced\":{\"restoreToOriginalReplicaCount\":true,\"horizontalPodAutoscalerConfig\":null},\"triggers\":[{\"type\":\"memory\",\"name\":\"\",\"metadata\":{\"type\":\"Utilization\",\"value\":\"20\"},\"authenticationRef\":null,\"metricType\":\"\"}],\"fallback\":{\"failureThreshold\":0,\"replicas\":1}}",
		IsApplied:              "Y",
	}
}

func generateRuntime() *dbclient.Runtime {
	return &dbclient.Runtime{
		BaseModel: dbengine.BaseModel{
			ID: 1,
		},
		Name:          "master",
		ApplicationID: 1,
		Workspace:     "PROD",
		GitBranch:     "",
		ProjectID:     1,
		Env:           "",
		ClusterName:   "test",
		ClusterId:     0,
		Creator:       "1",
		ScheduleName: dbclient.ScheduleName{
			Namespace: "services",
			Name:      "9ddcc8b369",
		},
		Status:              "Healthy",
		DeploymentStatus:    "",
		CurrentDeploymentID: 0,
		DeploymentOrderId:   "",
		ReleaseVersion:      "",
		LegacyStatus:        "",
		FileToken:           "",
		Deployed:            false,
		Deleting:            false,
		Version:             "",
		Source:              "",
		DiceVersion:         "",
		CPU:                 0.1,
		Mem:                 128,
		ConfigUpdatedDate:   nil,
		ReadableUniqueId:    "",
		GitRepoAbbrev:       "",
		OrgID:               1,
		ExtraParams:         "",
	}
}

func generatePreDeployment(tTime time.Time) *dbclient.PreDeployment {
	return &dbclient.PreDeployment{
		BaseModel: dbengine.BaseModel{
			ID:        1,
			CreatedAt: tTime,
			UpdatedAt: tTime,
		},
		ApplicationId: 1,
		Workspace:     "PROD",
		RuntimeName:   "master",
		Dice:          "{\"version\":\"2.0\",\"meta\":null,\"services\":{\"test\":{\"image\":\"addon-registry.default.svc.cluster.local:5000/erda-go-demo/go-web:go-demo-1645517335127632461\",\"image_username\":\"\",\"image_password\":\"\",\"cmd\":\"\",\"ports\":[{\"port\":8080,\"expose\":true}],\"resources\":{\"cpu\":0.1,\"mem\":128,\"max_cpu\":0,\"max_mem\":0,\"disk\":0,\"network\":{\"mode\":\"container\"}},\"deployments\":{\"replicas\":1,\"policies\":\"\"},\"health_check\":{\"http\":{},\"exec\":{}},\"traffic_security\":{}}}}",
		DiceOverlay:   "",
		DiceType:      1,
	}
}

func Test_podScalerService_GetRuntimeBaseInfo(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.ListRequest
	}
	tTime := time.Now()
	bis := make([]*pb.ServiceBaseInfo, 0)
	bis = append(bis, &pb.ServiceBaseInfo{
		ServiceName: "test",
		Deployments: &pb.Deployments{
			Replicas: 1,
		},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.RuntimeServiceBaseInfos
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ListRequest{
					RuntimeId: "1",
					Services:  "test",
				},
			},
			want: &pb.RuntimeServiceBaseInfos{
				RuntimeID:        1,
				ServiceBaseInfos: bis,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetPreDeployment",
				func(_ *dbServiceImpl, uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error) {
					return generatePreDeployment(tTime), nil
				})

			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.GetRuntimeBaseInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeBaseInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRuntimeBaseInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateRuntimeHPAEvents(tTime time.Time) []dbclient.HPAEventInfo {
	events := make([]dbclient.HPAEventInfo, 0)
	events = append(events, dbclient.HPAEventInfo{
		ID:          "1779dff5-184c-4dfe-9c76-978ac5126e59",
		CreatedAt:   tTime,
		UpdatedAt:   tTime,
		RuntimeID:   0,
		ServiceName: "test",
		Event:       "{\"lastTimestamp\":\"2022-06-27T05:56:53Z\",\"type\":\"Normal\",\"reason\":\"SuccessfulRescale\",\"message\":\"New size: 3; reason: memory resource utilization (percentage of request) above target\"}",
	})
	return events
}

func Test_podScalerService_ListRuntimeHPAEvents(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		ctx context.Context
		req *pb.ListRequest
	}

	tTime := time.Now()
	timestamppb.New(tTime).AsTime()
	events := make([]*pb.ErdaRuntimeHPAEvent, 0)
	events = append(events, &pb.ErdaRuntimeHPAEvent{
		ServiceName: "test",
		RuleId:      "1779dff5-184c-4dfe-9c76-978ac5126e59",
		Event: &pb.HPAEventDetail{
			CreateAt:     timestamppb.New(tTime),
			Type:         "Normal",
			Reason:       "SuccessfulRescale",
			EventMessage: "New size: 3; reason: memory resource utilization (percentage of request) above target",
		},
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ErdaRuntimeHPAEvents
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test_01",
			fields: fields{
				bundle:           &bundle.Bundle{},
				db:               &dbServiceImpl{},
				serviceGroupImpl: &servicegroup.ServiceGroupImpl{},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.ListRequest{
					RuntimeId: "1",
					Services:  "test",
				},
			},
			want:    &pb.ErdaRuntimeHPAEvents{Events: events},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetUserAndOrgID",
				func(_ *podscalerService, ctx context.Context) (userID user.ID, orgID uint64, err error) {
					return user.ID("1"), 1, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntime",
				func(_ *dbServiceImpl, id uint64) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.bundle), "CheckPermission",
				func(_ *bundle.Bundle, req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
					return &apistructs.PermissionCheckResponseData{
						Access: true,
					}, nil
				})
			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPAEventsByServices",
				func(_ *dbServiceImpl, runtimeId uint64, services []string) ([]dbclient.HPAEventInfo, error) {
					return generateRuntimeHPAEvents(tTime), nil
				})

			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.ListRuntimeHPAEvents(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRuntimeHPAEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListRuntimeHPAEvents() got = %v, want %v", got, tt.want)
			}
		})
	}
}
