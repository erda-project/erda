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

	"github.com/erda-project/erda-infra/providers/i18n"
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
