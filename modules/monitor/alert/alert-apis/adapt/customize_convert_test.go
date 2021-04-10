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
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"reflect"
	"testing"
)

func TestCustomizeAlertRule_ToModel(t *testing.T) {
	type fields struct {
		ID                  uint64
		Name                string
		Metric              string
		Window              uint64
		Functions           []*CustomizeAlertRuleFunction
		Filters             []*CustomizeAlertRuleFilter
		Group               []string
		Outputs             []string
		Select              map[string]string
		Attributes          map[string]interface{}
		ActivedMetricGroups []string
		CreateTime          int64
		UpdateTime          int64
	}
	type args struct {
		alert *CustomizeAlertDetail
		index string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *db.CustomizeAlertRule
	}{
		{
			name: "test_CustomizeAlertRule_ToModel",
			fields: fields{
				ID:        1,
				Name:      "test",
				Metric:    "metric",
				Window:    3,
				Functions: []*CustomizeAlertRuleFunction{},
				Filters:   []*CustomizeAlertRuleFilter{},
				Group:     []string{"group"},
				Outputs:   []string{"alert"},
			},
			args: args{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CustomizeAlertRule{
				ID:                  tt.fields.ID,
				Name:                tt.fields.Name,
				Metric:              tt.fields.Metric,
				Window:              tt.fields.Window,
				Functions:           tt.fields.Functions,
				Filters:             tt.fields.Filters,
				Group:               tt.fields.Group,
				Outputs:             tt.fields.Outputs,
				Select:              tt.fields.Select,
				Attributes:          tt.fields.Attributes,
				ActivedMetricGroups: tt.fields.ActivedMetricGroups,
				CreateTime:          tt.fields.CreateTime,
				UpdateTime:          tt.fields.UpdateTime,
			}
			if got := r.ToModel(tt.args.alert, tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
