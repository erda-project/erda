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
	"reflect"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"google.golang.org/protobuf/types/known/structpb"

	alertpb "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/internal/tools/monitor/core/alert/alert-apis/adapt"
)

type orgMock struct {
	mock.OrgMock
}

func (m orgMock) GetOrg(ctx context.Context, request *orgpb.GetOrgRequest) (*orgpb.GetOrgResponse, error) {
	if request.IdOrName == "" {
		return nil, fmt.Errorf("the IdOrName is empty")
	}
	if request.IdOrName != "1" {
		return nil, fmt.Errorf("org not found")
	}
	return &orgpb.GetOrgResponse{Data: &orgpb.Org{}}, nil
}

func Test_alertService_getOrg(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		orgIDOrName interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *orgpb.Org
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with error2",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: nil},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with no error",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: "1"},
			want:    &orgpb.Org{},
			wantErr: false,
		},
		{
			name: "test with StringValue",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: structpb.NewStringValue(strconv.Itoa(1))},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test with StringValue2",
			fields: fields{
				p: &provider{Org: orgMock{}},
			},
			args:    args{orgIDOrName: structpb.NewStringValue(strconv.Itoa(1)).AsInterface()},
			want:    &orgpb.Org{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &alertService{
				p: tt.fields.p,
			}
			got, err := m.GetOrg(tt.args.orgIDOrName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_alertService_CreateAlert(t *testing.T) {
	var a *adapt.Adapt

	monkey.PatchInstanceMethod(reflect.TypeOf(a), "CreateAlert", func(*adapt.Adapt, *alertpb.Alert) (alertID uint64, err error) {
		return 1, err
	})
	defer monkey.UnpatchAll()

	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *alertpb.CreateAlertRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *alertpb.CreateAlertResponse
		wantErr bool
	}{
		{
			name: "test with create alert",
			fields: fields{
				p: &provider{
					Org: orgMock{},
					a:   a,
				},
			},
			args: args{
				ctx: context.Background(),
				request: &alertpb.CreateAlertRequest{
					Id:           1,
					Name:         "test",
					AlertScope:   "org",
					AlertScopeId: "1",
					Enable:       false,
					Rules: []*alertpb.AlertExpression{
						{
							RuleId: 1,
						},
					},
					Notifies: []*alertpb.AlertNotify{
						{
							Id: 1,
						},
					},
					Filters: nil,
					Attributes: map[string]*structpb.Value{
						"dice_org_id": structpb.NewStringValue("1"),
					},
					ClusterNames:     nil,
					Domain:           "",
					CreateTime:       0,
					UpdateTime:       0,
					TriggerCondition: nil,
				},
			},
			want: &alertpb.CreateAlertResponse{
				Data: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &alertService{
				p: tt.fields.p,
			}
			got, err := m.CreateAlert(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAlert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAlert() got = %v, want %v", got, tt.want)
			}
		})
	}
}
