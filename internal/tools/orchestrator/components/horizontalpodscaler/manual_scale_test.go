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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/orchestrator/horizontalpodscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func Test_genOverlayDataForAudit(t *testing.T) {
	oldServiceData := &diceyml.Service{
		Resources: diceyml.Resources{
			CPU:  1,
			Mem:  1024,
			Disk: 0,
		},
		Deployments: diceyml.Deployments{
			Replicas: 1,
		},
	}

	auditData := genOverlayDataForAudit(oldServiceData)

	assert.Equal(t, float64(1), auditData.Resources.Cpu)
	assert.Equal(t, int64(1024), auditData.Resources.Mem)
	assert.Equal(t, int64(0), auditData.Resources.Disk)
	assert.Equal(t, uint64(1), auditData.Deployments.Replicas)
}

func Test_hpscalerService_processRuntimeScaleRecord(t *testing.T) {
	type fields struct {
		bundle           BundleService
		db               DBService
		serviceGroupImpl servicegroup.ServiceGroup
	}
	type args struct {
		rsc    pb.RuntimeScaleRecord
		action string
	}

	tTime := time.Now()

	services := make(map[string]*pb.RuntimeInspectServiceDTO)
	services["test"] = &pb.RuntimeInspectServiceDTO{
		Type:        "",
		Deployments: &pb.Deployments{Replicas: 2},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
	}

	wantsvc := make(map[string]*pb.RuntimeInspectServiceDTO)
	wantsvc["test"] = &pb.RuntimeInspectServiceDTO{
		Type:        "",
		Deployments: &pb.Deployments{Replicas: 1},
		Resources: &pb.Resources{
			Cpu:  0.1,
			Mem:  128,
			Disk: 0,
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.PreDiceDTO
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
				rsc: pb.RuntimeScaleRecord{
					ApplicationId: 1,
					Workspace:     "prod",
					Name:          "master",
					RuntimeId:     1,
					Payload: &pb.PreDiceDTO{
						Name:     "master",
						Envs:     nil,
						Services: services,
					},
					ErrorMsg: "",
				},
				action: "",
			},
			want: &pb.PreDiceDTO{
				Services: wantsvc,
			},
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
				rsc: pb.RuntimeScaleRecord{
					ApplicationId: 1,
					Workspace:     "prod",
					Name:          "master",
					RuntimeId:     1,
					Payload: &pb.PreDiceDTO{
						Name:     "master",
						Envs:     nil,
						Services: services,
					},
					ErrorMsg: "",
				},
				action: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &hpscalerService{
				bundle:           tt.fields.bundle,
				db:               tt.fields.db,
				serviceGroupImpl: tt.fields.serviceGroupImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeHPARulesByServices",
				func(_ *dbServiceImpl, id spec.RuntimeUniqueId, services []string) ([]dbclient.RuntimeHPA, error) {
					rules := make([]dbclient.RuntimeHPA, 0)
					rule := generateRuntimeHPA(tTime)
					rule.IsApplied = "N"
					if tt.name == "Test_02" {
						rule.IsApplied = "Y"
					}
					rule.ServiceName = "test"
					rule.Rules = "{\"ruleName\":\"test\",\"ruleNameSpace\":\"project-3-prod\",\"scaleTargetRef\":{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"name\":\"test-cdd1a9f7c4\",\"envSourceContainerName\":\"\"},\"pollingInterval\":0,\"cooldownPeriod\":0,\"minReplicaCount\":1,\"maxReplicaCount\":3,\"advanced\":{\"restoreToOriginalReplicaCount\":true,\"horizontalPodAutoscalerConfig\":null},\"triggers\":[{\"type\":\"memory\",\"name\":\"\",\"metadata\":{\"type\":\"Utilization\",\"value\":\"20\"},\"authenticationRef\":null,\"metricType\":\"\"}],\"fallback\":{\"failureThreshold\":0,\"replicas\":1}}"
					rules = append(rules, rule)
					return rules, nil
				})

			m2 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetPreDeployment",
				func(_ *dbServiceImpl, uniqueId spec.RuntimeUniqueId) (*dbclient.PreDeployment, error) {
					return generatePreDeployment(tTime), nil
				})

			m3 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetRuntimeByUniqueID",
				func(_ *dbServiceImpl, id spec.RuntimeUniqueId) (*dbclient.Runtime, error) {
					return generateRuntime(), nil
				})

			m4 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "GetUnDeletableAttachMentsByRuntimeID",
				func(_ *dbServiceImpl, runtimeID uint64) (*[]dbclient.AddonAttachment, error) {
					return nil, nil
				})

			m5 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "UpdateRuntime",
				func(_ *dbServiceImpl, runtime *dbclient.Runtime) error {
					return nil
				})

			m6 := monkey.PatchInstanceMethod(reflect.TypeOf(s.serviceGroupImpl), "Scale",
				func(_ *servicegroup.ServiceGroupImpl, sg *apistructs.ServiceGroup) (interface{}, error) {
					return nil, nil
				})
			m7 := monkey.PatchInstanceMethod(reflect.TypeOf(s.db), "UpdatePreDeployment",
				func(_ *dbServiceImpl, pre *dbclient.PreDeployment) error {
					return nil
				})

			defer m7.Unpatch()
			defer m6.Unpatch()
			defer m5.Unpatch()
			defer m4.Unpatch()
			defer m3.Unpatch()
			defer m2.Unpatch()
			defer m1.Unpatch()

			got, err := s.processRuntimeScaleRecord(tt.args.rsc, tt.args.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("processRuntimeScaleRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processRuntimeScaleRecord() got = %v, want %v", got, tt.want)
			}
		})
	}
}
