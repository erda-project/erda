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
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/monitor/utils"
	"reflect"
	"testing"
	"time"
)

func TestAlertRule_FromCustomizeAlertRule(t *testing.T) {
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
		cr   *db.CustomizeAlertRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *AlertRule
	}{
		{
			name:   "test_FromCustomizeAlertRule",
			fields: fields{},
			args: args{
				cr: &db.CustomizeAlertRule{
					ID:               12,
					Name:             "test_alert",
					CustomizeAlertID: 12,
					AlertType:        "alert",
					AlertIndex:       "dkfjkdjfk",
					AlertScope:       "app",
					AlertScopeID:     "18",
					Template:         utils.JSONMap{},
					Attributes:       utils.JSONMap{},
					Enable:           false,
					CreateTime:       time.Now(),
					UpdateTime:       time.Now(),
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
			if got := r.FromCustomizeAlertRule(tt.args.lang, tt.args.t, tt.args.cr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromCustomizeAlertRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
