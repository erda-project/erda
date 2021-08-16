// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package adapt

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/pkg/encoding/jsonmap"
)

type translator struct {
	common map[string]map[string]string
	dic    map[string]map[string]string
}

func (t *translator) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

func (t *translator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

func (t *translator) Get(lang i18n.LanguageCodes, key, def string) string {
	return def
}

func TestAlertRule_FromModel(t *testing.T) {
	type fields struct {
		ID         uint64
		Name       string
		AlertScope string
		AlertType  string
		AlertIndex *DisplayKey
		Template   map[string]interface{}
		Window     int64
		Functions  []*AlertRuleFunction
		IsRecover  bool
		Attributes map[string]interface{}
		Version    string
		Enable     bool
		CreateTime int64
		UpdateTime int64
	}
	type args struct {
		lang i18n.LanguageCodes
		t    i18n.Translator
		m    *db.AlertRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *AlertRule
	}{
		{
			name:   "test_AlertRule_FromModel",
			fields: fields{},
			args: args{
				lang: i18n.LanguageCodes{
					{
						Code: "zh",
					},
				},
				t: &translator{},
				m: &db.AlertRule{
					ID:         1,
					Name:       "test_alert",
					AlertScope: "app",
					AlertType:  "alert",
					AlertIndex: "dhfidjfkdfjd",
					Template: jsonmap.JSONMap{
						"window": 4,
						"functions": []interface{}{
							"str",
						},
					},
					Attributes: jsonmap.JSONMap{},
					Version:    "2",
					Enable:     false,
					CreateTime: time.Now(),
					UpdateTime: time.Now(),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AlertRule{
				ID:         tt.fields.ID,
				Name:       tt.fields.Name,
				AlertScope: tt.fields.AlertScope,
				AlertType:  tt.fields.AlertType,
				AlertIndex: tt.fields.AlertIndex,
				Template:   tt.fields.Template,
				Window:     tt.fields.Window,
				Functions:  tt.fields.Functions,
				IsRecover:  tt.fields.IsRecover,
				Attributes: tt.fields.Attributes,
				Version:    tt.fields.Version,
				Enable:     tt.fields.Enable,
				CreateTime: tt.fields.CreateTime,
				UpdateTime: tt.fields.UpdateTime,
			}
			if got := r.FromModel(tt.args.lang, tt.args.t, tt.args.m); got == nil {
				t.Errorf("FromModel() = nil")
			}
		})
	}
}

func TestAlertNotify_ToModel(t *testing.T) {
	type fields struct {
		ID          uint64
		Type        string
		GroupID     int64
		GroupType   string
		NotifyGroup *apistructs.NotifyGroup
		DingdingURL string
		Silence     *AlertNotifySilence
		CreateTime  int64
		UpdateTime  int64
	}
	type args struct {
		alert           *Alert
		silencePolicies map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *db.AlertNotify
	}{
		{
			name: "test_alertNotify_ToModel",
			fields: fields{
				ID:        33,
				Type:      "alert",
				GroupID:   12,
				GroupType: "dingding",
				NotifyGroup: &apistructs.NotifyGroup{
					ID: 12,
				},
				DingdingURL: "https://dingding",
				Silence: &AlertNotifySilence{
					Value: 5,
					Unit:  "seconds",
				},
				CreateTime: 0,
				UpdateTime: 0,
			},
			args: args{
				alert: &Alert{
					ID:     3,
					Enable: true,
				},
				silencePolicies: map[string]bool{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &AlertNotify{
				ID:          tt.fields.ID,
				Type:        tt.fields.Type,
				GroupID:     tt.fields.GroupID,
				GroupType:   tt.fields.GroupType,
				NotifyGroup: tt.fields.NotifyGroup,
				DingdingURL: tt.fields.DingdingURL,
				Silence:     tt.fields.Silence,
				CreateTime:  tt.fields.CreateTime,
				UpdateTime:  tt.fields.UpdateTime,
			}
			if got := n.ToModel(tt.args.alert, tt.args.silencePolicies); got == nil {
				t.Errorf("ToModel() = nil")
			}
		})
	}
}

func TestToDBAlertExpressionModel(t *testing.T) {
	type args struct {
		e       *pb.AlertExpression
		orgName string
		alert   *pb.Alert
		rule    *pb.AlertRule
	}
	clusterName, _ := structpb.NewList([]interface{}{
		"terminus-test",
		"fdp-test",
	})
	templateCluster, _ := structpb.NewList([]interface{}{
		"cluster_name",
	})
	templateOutput, _ := structpb.NewList([]interface{}{
		"alert",
	})
	ruleAttribute, _ := structpb.NewList([]interface{}{
		"machine_mem",
		"machine",
	})
	templateFilters, _ := structpb.NewList([]interface{}{
		map[string]*structpb.Value{
			"tag":      structpb.NewStringValue("cluster_name"),
			"operator": structpb.NewStringValue("neq"),
			"value":    structpb.NewStringValue("xxxxxxx"),
			"dataType": structpb.NewStringValue(""),
		},
		map[string]*structpb.Value{
			"tag":      structpb.NewStringValue("cluster_name"),
			"operator": structpb.NewStringValue("in"),
			"value":    structpb.NewStringValue("$cluster_name"),
			"dataType": structpb.NewStringValue(""),
		},
		map[string]*structpb.Value{
			"tag":      structpb.NewStringValue("_metric_scope"),
			"operator": structpb.NewStringValue("eq"),
			"value":    structpb.NewStringValue("org"),
			"dataType": structpb.NewStringValue(""),
		},
		map[string]*structpb.Value{
			"tag":      structpb.NewStringValue("_metric_scope_id"),
			"operator": structpb.NewStringValue("eq"),
			"value":    structpb.NewStringValue("terminus"),
			"dataType": structpb.NewStringValue(""),
		},
	})
	templateFunctions, _ := structpb.NewList([]interface{}{
		map[string]*structpb.Value{
			"unit":       structpb.NewStringValue(""),
			"field":      structpb.NewStringValue("mem_used"),
			"alias":      structpb.NewStringValue("sdf"),
			"aggregator": structpb.NewStringValue("sum"),
			"operator":   structpb.NewStringValue("neq"),
			"value":      structpb.NewNumberValue(float64(1)),
			"dataType":   structpb.NewStringValue(""),
		},
	})
	templateSelect, _ := structpb.NewList([]interface{}{
		map[string]*structpb.Value{
			"os":       structpb.NewStringValue("#os"),
			"hostname": structpb.NewStringValue("#hostname"),
		},
	})

	tests := []struct {
		name    string
		args    args
		want    *db.AlertExpression
		wantErr bool
	}{
		{
			name: "TestToDBAlertExpressionModel",
			args: args{
				e: &pb.AlertExpression{
					Id:         44,
					RuleId:     0,
					AlertIndex: "96683ebf-a1de-4756-9f55-e2794e00a19e",
					Window:     1,
					Functions: []*pb.AlertExpressionFunction{
						{
							Field:      "mem_used",
							Aggregator: "sum",
							Operator:   "neq",
							Value:      structpb.NewNumberValue(1),
						},
					},
					IsRecover:  false,
					CreateTime: 1628821079000,
					UpdateTime: 1628821079000,
				},
				orgName: "terminus",
				alert: &pb.Alert{
					Id:           56,
					Name:         "pjycccc",
					AlertScope:   "org",
					AlertScopeId: "1",
					Enable:       true,
					Rules: []*pb.AlertExpression{
						{
							Id:         44,
							RuleId:     0,
							AlertIndex: "96683ebf-a1de-4756-9f55-e2794e00a19e",
							Window:     1,
							Functions: []*pb.AlertExpressionFunction{
								{
									Field:      "mem_used",
									Aggregator: "sum",
									Operator:   "neq",
									Value:      structpb.NewNumberValue(1),
								},
							},
							IsRecover:  false,
							CreateTime: 1628821079000,
							UpdateTime: 1628821079000,
						},
					},
					Notifies: []*pb.AlertNotify{
						{
							Id:          0,
							Type:        "notify_group",
							GroupId:     21,
							GroupType:   "mbox",
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
						"alert_domain":         structpb.NewStringValue("https://erda.test.terminus.io"),
						"alert_dashboard_path": structpb.NewStringValue("/dataCenter/customDashboard"),
						"alert_record_path":    structpb.NewStringValue("/dataCenter/alarm/record"),
						"dice_org_id":          structpb.NewStringValue("1"),
						"cluster_name":         structpb.NewListValue(clusterName),
					},
					ClusterNames: []string{"terminus-test", "fdp-test"},
					Domain:       "https://erda.test.terminus.io",
					CreateTime:   0,
					UpdateTime:   0,
				},
				rule: &pb.AlertRule{
					Id:         44,
					Name:       "erda_test",
					AlertScope: "org",
					AlertType:  "org_customize",
					AlertIndex: &pb.DisplayKey{
						Key:     "96683ebf-a1de-4756-9f55-e2794e00a19e",
						Display: "erda_test",
					},
					Template: map[string]*structpb.Value{
						"filters":   structpb.NewListValue(templateFilters),
						"window":    structpb.NewNumberValue(float64(1)),
						"functions": structpb.NewListValue(templateFunctions),
						"group":     structpb.NewListValue(templateCluster),
						"metric":    structpb.NewStringValue("host_summary"),
						"outputs":   structpb.NewListValue(templateOutput),
						"select":    structpb.NewListValue(templateSelect),
					},
					Window: 1,
					Functions: []*pb.AlertRuleFunction{
						{
							Field: &pb.DisplayKey{
								Key:     "mem_used",
								Display: "mem_used",
							},
							Aggregator: "sum",
							Operator:   "neq",
							Value:      structpb.NewNumberValue(float64(1)),
							DataType:   "number",
							Unit:       "",
						},
					},
					IsRecover: false,
					Attributes: map[string]*structpb.Value{
						"active_metric_groups": structpb.NewListValue(ruleAttribute),
						"alert_dashboard_id":   structpb.NewStringValue("e1730156951c47ec90ea79552b596a39"),
						"alert_group":          structpb.NewStringValue("{{cluster_name}}-{{cluster_name}}"),
						"level":                structpb.NewStringValue("WARNING"),
						"recover":              structpb.NewStringValue("false"),
						"tickets_metric_key":   structpb.NewStringValue("{{cluster_name}}"),
					},
					Version:    "",
					Enable:     true,
					CreateTime: 1628821079000,
					UpdateTime: 1628821079000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDBAlertExpressionModel(tt.args.e, tt.args.orgName, tt.args.alert, tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDBAlertExpressionModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ToDBAlertExpressionModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}
