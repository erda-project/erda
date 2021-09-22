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

package adapt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/encoding/jsonmap"
)

func TestAdapt_newTicketAlertNotify(t *testing.T) {
	type fields struct {
		l logs.Logger
		//metricq                Queryer
		t    i18n.Translator
		db   *db.DB
		cql  *cql.Cql
		bdl  *bundle.Bundle
		cmdb *cmdb.Cmdb
		//dashboardAPI           DashboardAPI
		orgFilterTags          map[string]bool
		microServiceFilterTags map[string]bool
		silencePolicies        map[string]bool
	}
	type args struct {
		alertID uint64
		silence *pb.AlertNotifySilence
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *db.AlertNotify
	}{
		//{
		//	name: "test_newTicketAlertNotify",
		//	fields: fields{
		//		silencePolicies: map[string]bool{
		//			"silence": true,
		//		},
		//	},
		//	args: args{
		//		alertID: 11,
		//		silence: &AlertNotifySilence{
		//			Value:  5,
		//			Unit:   "second",
		//			Policy: "silence",
		//		},
		//	},
		//	want: nil,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapt{
				l: tt.fields.l,
				//metricq:                tt.fields.metricq,
				t:    tt.fields.t,
				db:   tt.fields.db,
				cql:  tt.fields.cql,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
				//dashboardAPI:           tt.fields.dashboardAPI,
				orgFilterTags:          tt.fields.orgFilterTags,
				microServiceFilterTags: tt.fields.microServiceFilterTags,
				silencePolicies:        tt.fields.silencePolicies,
			}
			if got := a.newTicketAlertNotify(tt.args.alertID, tt.args.silence); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTicketAlertNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapt_compareNotify(t *testing.T) {
	type fields struct {
		l logs.Logger
		//metricq                Queryer
		t    i18n.Translator
		db   *db.DB
		cql  *cql.Cql
		bdl  *bundle.Bundle
		cmdb *cmdb.Cmdb
		//dashboardAPI           DashboardAPI
		orgFilterTags          map[string]bool
		microServiceFilterTags map[string]bool
		silencePolicies        map[string]bool
	}
	type args struct {
		a *db.AlertNotify
		b *db.AlertNotify
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test_compareNotify",
			fields: fields{},
			args: args{
				a: &db.AlertNotify{
					NotifyTarget: jsonmap.JSONMap{},
				},
				b: &db.AlertNotify{
					NotifyTarget: jsonmap.JSONMap{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &Adapt{
				l: tt.fields.l,
				//metricq:                tt.fields.metricq,
				t:    tt.fields.t,
				db:   tt.fields.db,
				cql:  tt.fields.cql,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
				//dashboardAPI:           tt.fields.dashboardAPI,
				orgFilterTags:          tt.fields.orgFilterTags,
				microServiceFilterTags: tt.fields.microServiceFilterTags,
				silencePolicies:        tt.fields.silencePolicies,
			}
			if got := ad.compareNotify(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("compareNotify() = %v, want %v", got, tt.want)
			}
		})
	}
}

//go:generate mockgen -destination=./alert_logs_test.go -package adapt github.com/erda-project/erda-infra/base/logs Logger
//go:generate mockgen -destination=./alert_metricq_test.go -package adapt github.com/erda-project/erda/modules/core/monitor/metric/query/metricq Queryer
//go:generate mockgen -destination=./alert_t_test.go -package adapt github.com/erda-project/erda-infra/providers/i18n Translator

type pLog struct {
}

func (p *pLog) Print(values ...interface{}) {
	fmt.Println("ok")
}
func TestAdapt_UpdateOrgAlert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	defer monkey.UnpatchAll()
	logsss := NewMockLogger(ctrl)
	logsss.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes().Do(fmt.Errorf("err"))

	pdb := &gorm.DB{}
	pdb.SetLogger(&pLog{})
	monkey.Patch((*db.DB).Begin, func(_ *db.DB) *db.DB {
		return &db.DB{
			DB: pdb,
			CustomizeAlert: db.CustomizeAlertDB{
				pdb,
			},
			CustomizeAlertRule: db.CustomizeAlertRuleDB{
				pdb,
			},
			CustomizeAlertNotifyTemplate: db.CustomizeAlertNotifyTemplateDB{
				pdb,
			},
			Alert: db.AlertDB{
				pdb,
			},
			AlertExpression: db.AlertExpressionDB{
				pdb,
			},
			AlertNotify: db.AlertNotifyDB{
				pdb,
			},
			AlertNotifyTemplate: db.AlertNotifyTemplateDB{
				pdb,
			},
			AlertRule: db.AlertRuleDB{
				pdb,
			},
			AlertRecord: db.AlertRecordDB{
				pdb,
			},
		}
	})
	monkey.Patch((*db.AlertDB).GetByScopeAndScopeIDAndName, func(_ *db.AlertDB, _, _, _ string) (*db.Alert, error) {
		return &db.Alert{
			ID:           20,
			Name:         "wwwwwww",
			AlertScope:   "micro_service",
			AlertScopeID: "163334b65f7a4f504d8ca11733d71ea7",
			Attributes: jsonmap.JSONMap{
				"alert_domain":          "https://erda.dev.terminus.io",
				"alert_record_path":     "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/alarm-record",
				"application_id":        []string{"16"},
				"project_id":            []string{"14"},
				"target_project_id":     []string{"14"},
				"tenant_group":          "163334b65f7a4f504d8ca11733d71ea7",
				"tk":                    "163334b65f7a4f504d8ca11733d71ea7",
				"alert_dashboard_path":  "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/custom-dashboard",
				"dice_org_id":           "2",
				"target_application_id": []string{"16"},
				"target_workspace":      "DEV",
				"terminus_key":          "163334b65f7a4f504d8ca11733d71ea7",
				"workspace":             "DEV",
			},
			Enable:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}, nil
	})
	monkey.Patch((*db.AlertDB).GetByID, func(_ *db.AlertDB, _ uint64) (*db.Alert, error) {
		return &db.Alert{
			ID:           20,
			Name:         "wwwwwww",
			AlertScope:   "micro_service",
			AlertScopeID: "163334b65f7a4f504d8ca11733d71ea7",
			Attributes: jsonmap.JSONMap{
				"alert_domain":          "https://erda.dev.terminus.io",
				"alert_record_path":     "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/alarm-record",
				"application_id":        []interface{}{"16"},
				"project_id":            []interface{}{"14"},
				"target_project_id":     []interface{}{"14"},
				"tenant_group":          "163334b65f7a4f504d8ca11733d71ea7",
				"tk":                    "163334b65f7a4f504d8ca11733d71ea7",
				"alert_dashboard_path":  "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/custom-dashboard",
				"dice_org_id":           "2",
				"target_application_id": []interface{}{"16"},
				"target_workspace":      "DEV",
				"terminus_key":          "163334b65f7a4f504d8ca11733d71ea7",
				"workspace":             "DEV",
			},
			Enable:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}, nil
	})
	monkey.Patch((*db.AlertDB).Update, func(_ *db.AlertDB, _ *db.Alert) error {
		return nil
	})
	monkey.Patch(ToDBAlertModel, func(_ *pb.Alert) *db.Alert {
		return &db.Alert{
			ID:           20,
			Name:         "wwwwwww",
			AlertScope:   "micro_service",
			AlertScopeID: "163334b65f7a4f504d8ca11733d71ea7",
			Attributes: jsonmap.JSONMap{
				"alert_domain":          "https://erda.dev.terminus.io",
				"alert_record_path":     "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/alarm-record",
				"application_id":        "16",
				"project_id":            "14",
				"target_project_id":     "14",
				"tenant_group":          "163334b65f7a4f504d8ca11733d71ea7",
				"tk":                    "163334b65f7a4f504d8ca11733d71ea7",
				"alert_dashboard_path":  "/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/alarm-management/163334b65f7a4f504d8ca11733d71ea7/custom-dashboard",
				"dice_org_id":           "2",
				"target_application_id": "16",
				"target_workspace":      "DEV",
				"terminus_key":          "163334b65f7a4f504d8ca11733d71ea7",
				"workspace":             "DEV",
			},
			Enable:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}
	})
	monkey.Patch((*Adapt).getEnabledAlertRulesByScopeAndIndices, func(_ *Adapt, _ i18n.LanguageCodes, _, _ string, _ []string) (map[string]*pb.AlertRule, error) {
		return map[string]*pb.AlertRule{
			"app_resource_container_cpu": &pb.AlertRule{
				Id:         0,
				Name:       "",
				AlertScope: "",
				AlertType:  "",
				AlertIndex: &pb.DisplayKey{
					Key:     "app_resource_container_cpu",
					Display: "应用实例CPU使用率异常",
				},
				Template:   nil,
				Window:     0,
				Functions:  nil,
				IsRecover:  false,
				Attributes: nil,
				Version:    "",
				Enable:     false,
				CreateTime: 0,
				UpdateTime: 0,
			},
		}, nil
	})
	monkey.Patch((*Adapt).getAlertExpressionsMapByAlertID, func(_ *Adapt, _ uint64) (map[uint64]*db.AlertExpression, error) {
		return map[uint64]*db.AlertExpression{
			29: {
				ID:      29,
				AlertID: 20,
				Attributes: jsonmap.JSONMap{
					"alert_id":       "20",
					"alert_scope_id": "163334b65f7a4f504d8ca11733d71ea7",
					"alert_type":     "app_resource",
				},
				Expression: jsonmap.JSONMap{
					"metric": "docker_container_summary",
					"window": 1,
				},
				Version: "3.0",
				Enable:  true,
				Created: time.Now(),
				Updated: time.Now(),
			},
		}, nil
	})
	monkey.Patch(ToDBAlertExpressionModel, func(_ *pb.AlertExpression, _ string, _ *pb.Alert, _ *pb.AlertRule) (*db.AlertExpression, error) {
		return &db.AlertExpression{
			ID:      29,
			AlertID: 20,
			Attributes: jsonmap.JSONMap{
				"alert_id":       "20",
				"alert_scope_id": "163334b65f7a4f504d8ca11733d71ea7",
				"alert_type":     "app_resource",
			},
			Expression: jsonmap.JSONMap{
				"metric": "docker_container_summary",
				"window": 1,
			},
			Version: "3.0",
			Enable:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}, nil
	})
	monkey.Patch((*db.AlertExpressionDB).Update, func(_ *db.AlertExpressionDB, _ *db.AlertExpression) error {
		return nil
	})
	monkey.Patch((*db.AlertExpressionDB).Insert, func(_ *db.AlertExpressionDB, _ *db.AlertExpression) error {
		return nil
	})
	monkey.Patch((*db.AlertExpressionDB).DeleteByIDs, func(_ *db.AlertExpressionDB, ids []uint64) error {
		return nil
	})
	monkey.Patch((*db.AlertNotifyDB).DeleteByIDs, func(_ *db.AlertNotifyDB, _ []uint64) error {
		return nil
	})
	monkey.Patch((*Adapt).getEnabledAlertRulesByScopeAndIndices, func(_ *Adapt, lang i18n.LanguageCodes, scope, scopeID string, indices []string) (map[string]*pb.AlertRule, error) {
		return map[string]*pb.AlertRule{
			"app_resource_container_cpu": {
				Id:         788,
				Name:       "应用实例CPU使用率异常",
				AlertScope: "micro_service",
				AlertType:  "app_resource",
				AlertIndex: &pb.DisplayKey{},
				Template:   nil,
				Window:     1,
				Functions:  make([]*pb.AlertRuleFunction, 0),
				IsRecover:  true,
				Attributes: nil,
				Version:    "",
				Enable:     true,
				CreateTime: 1629306463000,
				UpdateTime: 1629306463000,
			},
		}, nil
	})
	monkey.Patch((*Adapt).getAlertExpressionsMapByAlertID, func(_ *Adapt, alertID uint64) (map[uint64]*db.AlertExpression, error) {
		return map[uint64]*db.AlertExpression{
			29: {
				ID:         29,
				AlertID:    29,
				Attributes: nil,
				Expression: nil,
				Version:    "3.0",
				Enable:     true,
				Created:    time.Now(),
				Updated:    time.Now(),
			},
		}, nil
	})
	monkey.Patch(ToDBAlertExpressionModel, func(e *pb.AlertExpression, orgName string, alert *pb.Alert, rule *pb.AlertRule) (*db.AlertExpression, error) {
		return &db.AlertExpression{
			ID:      29,
			AlertID: 20,
			Attributes: jsonmap.JSONMap{
				"target_project_id": "14",
				"workspace":         "DEV",
				"display_url":       "https://erda.dev.terminus.io/erda/workBench/projects/14/apps/{{application_id}}/deploy/runtimes/{{runtime_id}}/overview",
			},
			Expression: jsonmap.JSONMap{
				"metric":  "docker_container_summary",
				"outputs": []string{"alert"},
			},
			Version: "3.0",
			Enable:  true,
			Created: time.Now(),
			Updated: time.Now(),
		}, nil
	})
	monkey.Patch((*Adapt).getAlertNotifysMapByAlertID, func(_ *Adapt, alertID uint64) (map[uint64]*db.AlertNotify, error) {
		return map[uint64]*db.AlertNotify{
			48: {
				ID:        48,
				AlertID:   20,
				NotifyKey: "",
				NotifyTarget: jsonmap.JSONMap{
					"type": "ticket",
				},
				NotifyTargetID: "",
				Silence:        300000,
				SilencePolicy:  "fixed",
				Enable:         true,
				Created:        time.Now(),
				Updated:        time.Now(),
			},
			49: {
				ID:        49,
				AlertID:   20,
				NotifyKey: "",
				NotifyTarget: jsonmap.JSONMap{
					"group_id":   21,
					"group_type": "mbox,email",
					"type":       "notify_group",
				},
				NotifyTargetID: "",
				Silence:        300000,
				SilencePolicy:  "fixed",
				Enable:         true,
				Created:        time.Now(),
				Updated:        time.Now(),
			},
		}, nil
	})
	monkey.Patch((*db.AlertNotifyDB).Update, func(_ *db.AlertNotifyDB, _ *db.AlertNotify) error {
		return nil
	})

	ada := &Adapt{
		l:       logsss,
		metricq: NewMockQueryer(ctrl),
		t:       NewMockTranslator(ctrl),
		db: &db.DB{
			DB: pdb,
			CustomizeAlert: db.CustomizeAlertDB{
				pdb,
			},
			CustomizeAlertRule: db.CustomizeAlertRuleDB{
				pdb,
			},
			CustomizeAlertNotifyTemplate: db.CustomizeAlertNotifyTemplateDB{
				pdb,
			},
			Alert: db.AlertDB{
				pdb,
			},
			AlertExpression: db.AlertExpressionDB{
				pdb,
			},
			AlertNotify: db.AlertNotifyDB{
				pdb,
			},
			AlertNotifyTemplate: db.AlertNotifyTemplateDB{
				pdb,
			},
			AlertRule: db.AlertRuleDB{
				pdb,
			},
			AlertRecord: db.AlertRecordDB{
				pdb,
			},
		},
		cql:                         &cql.Cql{},
		bdl:                         &bundle.Bundle{},
		cmdb:                        &cmdb.Cmdb{},
		dashboardAPI:                nil,
		orgFilterTags:               nil,
		microServiceFilterTags:      nil,
		microServiceOtherFilterTags: nil,
		silencePolicies:             nil,
	}

	err := ada.UpdateAlert(20, &pb.Alert{
		Id:           20,
		Name:         "wwwwwww",
		AlertScope:   "micro_service",
		AlertScopeId: "163334b65f7a4f504d8ca11733d71ea7",
		Enable:       false,
		Rules: []*pb.AlertExpression{
			{
				Id:         29,
				RuleId:     788,
				AlertIndex: "app_resource_container_cpu",
				Window:     1,
				Functions: []*pb.AlertExpressionFunction{
					{
						Field:      "cpu_usage_percent",
						Aggregator: "avg",
						Operator:   "neq",
						Value:      structpb.NewNumberValue(float64(90)),
					},
				},
				IsRecover:  true,
				CreateTime: 1632138875,
				UpdateTime: 1632142281,
			},
		},
		Notifies: []*pb.AlertNotify{
			{
				Id:          0,
				Type:        "notify_group",
				GroupId:     21,
				GroupType:   "mbox,email",
				NotifyGroup: nil,
				DingdingUrl: "",
				Silence: &pb.AlertNotifySilence{
					Value:  5,
					Unit:   "minutes",
					Policy: "fixed",
				},
				CreateTime: 0,
				UpdateTime: 0,
			},
		},
		Filters: nil,
		Attributes: map[string]*structpb.Value{
			"alert_dashboard_path": structpb.NewStringValue("/microService/14/DEV/163334b65f7a4f504d8ca11733d71ea7/monitor/163334b65f7a4f504d8ca11733d71ea7/custom-dashboard"),
			"alert_domain":         structpb.NewStringValue("https://erda.dev.terminus.io"),
			"application_id":       structpb.NewStringValue("16"),
			"org_name":             structpb.NewStringValue("terminus"),
		},
		ClusterNames: nil,
		Domain:       "https://erda.dev.terminus.io",
		CreateTime:   0,
		UpdateTime:   0,
	})
	if err != nil {
		fmt.Println("should not err")
	}
}
