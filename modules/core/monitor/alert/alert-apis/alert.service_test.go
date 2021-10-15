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

package apis

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"bou.ke/monkey"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/common/apis"
)

func Test_alertService_GetAlertConditions(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetAlertConditionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetAlertConditionsResponse
		wantErr bool
	}{
		{
			name: "test_getAlertConditions",
			fields: fields{
				p: &provider{
					C: &config{
						OrgFilterTags:               "",
						MicroServiceFilterTags:      "",
						MicroServiceOtherFilterTags: "",
						SilencePolicy:               "",
						AlertConditions:             "./../../../../../conf/monitor/monitor/alert/trigger_conditions.yaml",
						Cassandra: struct {
							cassandra.SessionConfig `file:"session"`
							GCGraceSeconds          int `file:"gc_grace_seconds" default:"86400"`
						}{},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				request: &pb.GetAlertConditionsRequest{
					ScopeType: Org,
				},
			},
			wantErr: false,
		},
		{
			name: "test_getAlertConditions",
			fields: fields{
				p: &provider{
					C: &config{
						OrgFilterTags:               "",
						MicroServiceFilterTags:      "",
						MicroServiceOtherFilterTags: "",
						SilencePolicy:               "",
						AlertConditions:             "./../../../../../conf/monitor/monitor/alert/trigger_conditions.yaml",
						Cassandra: struct {
							cassandra.SessionConfig `file:"session"`
							GCGraceSeconds          int `file:"gc_grace_seconds" default:"86400"`
						}{},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				request: &pb.GetAlertConditionsRequest{
					ScopeType: "test",
				},
			},
			wantErr: false,
		},
	}
	tests[0].fields.p.alertConditions = make([]*AlertConditions, 0)
	f, err := ioutil.ReadFile(tests[0].fields.p.C.AlertConditions)
	if err != nil {
		fmt.Println("read file is fail ", err)
	}
	err = yaml.Unmarshal(f, &tests[0].fields.p.alertConditions)
	if err != nil {
		fmt.Println("unmarshal is fail ", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &alertService{
				p: tt.fields.p,
			}
			_, err := m.GetAlertConditions(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAlertConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

////go:generate mockgen -destination=./condition_metric_test.go -package apis github.com/erda-project/erda-proto-go/core/monitor/metric/pb MetricServiceServer
func Test_alertService_GetAlertConditionsValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	defer monkey.UnpatchAll()
	metricService := NewMockMetricServiceServer(ctrl)
	metricService.EXPECT().QueryWithInfluxFormat(gomock.Any(), gomock.Any()).AnyTimes().Return(&metricpb.QueryWithInfluxFormatResponse{
		Results: []*metricpb.Result{
			{
				Series: []*metricpb.Serie{
					{
						Rows: []*metricpb.Row{
							{
								Values: []*structpb.Value{
									structpb.NewStringValue("go-demo"),
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	monkey.Patch(apis.GetOrgID, func(_ context.Context) string {
		return "2"
	})
	monkey.Patch((*bundle.Bundle).GetOrg, func(_ *bundle.Bundle, _ interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID:   1,
			Name: "erda",
		}, nil
	})
	pro := &provider{
		C: &config{
			AlertConditions: "./../../../../../conf/monitor/monitor/alert/trigger_conditions.yaml",
		},
		alertConditions: make([]*AlertConditions, 0),
		alertService:    &alertService{},
		Metric:          metricService,
	}

	f, err := ioutil.ReadFile(pro.C.AlertConditions)
	if err != nil {
		fmt.Println("read file is fail ", err)
	}
	err = yaml.Unmarshal(f, &pro.alertConditions)
	if err != nil {
		fmt.Println("unmarshal is fail ", err)
	}

	pro.alertService.p = pro
	_, err = pro.alertService.GetAlertConditionsValue(context.Background(), &pb.GetAlertConditionsValueRequest{
		ProjectId:   "3",
		TerminusKey: "sg44yfh5464g6uy56j7224f",
		ScopeType:   "org",
	})
	_, err = pro.alertService.GetAlertConditionsValue(context.Background(), &pb.GetAlertConditionsValueRequest{
		ProjectId:   "3",
		TerminusKey: "sg44yfh5464g6uy56j7224f",
		ScopeType:   "msp",
	})
	if err != nil {
		fmt.Println("should not err")
	}
}
