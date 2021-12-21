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

package expression

import (
	"context"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
	alertdb "github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/pkg/encoding/jsonmap"
)

func Test_expressionService_GetAllEnabledExpression(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetAllAlertEnabledExpressionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test case",
			fields: fields{
				p: &provider{
					Cfg:               nil,
					Log:               nil,
					Register:          nil,
					DB:                nil,
					alertDB:           &alertdb.AlertExpressionDB{&gorm.DB{}},
					metricDB:          &alertdb.MetricExpressionDB{&gorm.DB{}},
					expressionService: nil,
				},
			},
			args: args{
				ctx: nil,
				request: &pb.GetAllAlertEnabledExpressionRequest{
					PageSize: 1,
					PageNo:   1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var adb *alertdb.AlertExpressionDB
			alertExpression := monkey.PatchInstanceMethod(reflect.TypeOf(adb), "GetAllAlertExpression", func(db *alertdb.AlertExpressionDB) ([]*alertdb.AlertExpression, error) {
				return []*alertdb.AlertExpression{
					{
						ID:         1,
						AlertID:    1,
						Attributes: jsonmap.JSONMap{"org_name": "erda"},
						Expression: jsonmap.JSONMap{"outputs": []string{"alert"}},
						Version:    "1.0",
						Enable:     true,
						Created:    time.Time{},
						Updated:    time.Time{},
					},
				}, nil
			})
			defer alertExpression.Unpatch()
			var mdb *alertdb.MetricExpressionDB
			metricExpression := monkey.PatchInstanceMethod(reflect.TypeOf(mdb), "GetAllMetricExpression", func(db *alertdb.MetricExpressionDB) ([]*alertdb.MetricExpression, error) {
				return []*alertdb.MetricExpression{
					{
						ID:         1,
						Attributes: nil,
						Expression: jsonmap.JSONMap{
							"window": 3,
						},
						Version: "1.0",
						Enable:  true,
						Created: time.Time{},
						Updated: time.Time{},
					},
				}, nil
			})
			defer metricExpression.Unpatch()
			e := &expressionService{
				p: tt.fields.p,
			}
			_, err := e.GetAllAlertEnabledExpression(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllEnabledExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_expressionService_GetAllAlertTemplate(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetAllAlertTemplateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.GetAllAlertTemplateResponse
		wantErr bool
	}{
		{
			name: "test case",
			fields: fields{
				p: &provider{
					Cfg:                  nil,
					Log:                  nil,
					Register:             nil,
					t:                    &i18n.NopTranslator{},
					DB:                   nil,
					alertDB:              &alertdb.AlertExpressionDB{&gorm.DB{}},
					metricDB:             &alertdb.MetricExpressionDB{&gorm.DB{}},
					customizeAlertRuleDB: &alertdb.CustomizeAlertRuleDB{&gorm.DB{}},
					expressionService:    nil,
				},
			},
			args: args{
				ctx: nil,
				request: &pb.GetAllAlertTemplateRequest{
					PageSize: 1,
					PageNo:   1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &expressionService{
				p: tt.fields.p,
			}
			_, err := e.GetAllAlertTemplate(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllAlertTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_expressionService_GetAllMetricEnabledExpression(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx     context.Context
		request *pb.GetAllMetricEnabledExpressionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test case",
			fields: fields{
				p: &provider{
					Cfg:                  nil,
					Log:                  nil,
					Register:             nil,
					t:                    &i18n.NopTranslator{},
					DB:                   nil,
					alertDB:              &alertdb.AlertExpressionDB{&gorm.DB{}},
					metricDB:             &alertdb.MetricExpressionDB{&gorm.DB{}},
					customizeAlertRuleDB: &alertdb.CustomizeAlertRuleDB{&gorm.DB{}},
					expressionService:    nil,
				},
			},
			args: args{
				ctx: nil,
				request: &pb.GetAllMetricEnabledExpressionRequest{
					PageSize: 1,
					PageNo:   1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		var adb *alertdb.AlertExpressionDB
		alertExpression := monkey.PatchInstanceMethod(reflect.TypeOf(adb), "GetAllAlertExpression", func(db *alertdb.AlertExpressionDB) ([]*alertdb.AlertExpression, error) {
			return []*alertdb.AlertExpression{
				{
					ID:         1,
					AlertID:    1,
					Attributes: jsonmap.JSONMap{"org_name": "erda"},
					Expression: jsonmap.JSONMap{"outputs": []string{"alert"}},
					Version:    "1.0",
					Enable:     true,
					Created:    time.Time{},
					Updated:    time.Time{},
				},
			}, nil
		})
		defer alertExpression.Unpatch()
		var mdb *alertdb.MetricExpressionDB
		metricExpression := monkey.PatchInstanceMethod(reflect.TypeOf(mdb), "GetAllMetricExpression", func(db *alertdb.MetricExpressionDB) ([]*alertdb.MetricExpression, error) {
			return []*alertdb.MetricExpression{
				{
					ID:         1,
					Attributes: nil,
					Expression: jsonmap.JSONMap{
						"window": 3,
					},
					Version: "1.0",
					Enable:  true,
					Created: time.Time{},
					Updated: time.Time{},
				},
			}, nil
		})
		defer metricExpression.Unpatch()
		t.Run(tt.name, func(t *testing.T) {
			e := &expressionService{
				p: tt.fields.p,
			}
			_, err := e.GetAllMetricEnabledExpression(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllMetricEnabledExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
