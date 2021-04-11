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

package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/erda-project/erda/apistructs"
)

func TestNotifyGroup_ToApiData(t *testing.T) {
	type fields struct {
		BaseModel   BaseModel
		Name        string
		ScopeType   string
		ScopeID     string
		OrgID       int64
		TargetData  string
		Label       string
		ClusterName string
		AutoCreate  bool
		Creator     string
	}
	tests := []struct {
		name   string
		fields fields
		want   *apistructs.NotifyGroup
	}{
		{
			name: "test_ToApiData",
			fields: fields{
				BaseModel: BaseModel{
					ID:        14,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:       "test",
				ScopeType:  "app",
				ScopeID:    "18",
				OrgID:      1,
				TargetData: `[{"type":"dingding","values":[{"receiver":"https://oapi.dingtalk.com","secret":"e69d6e340a07c47fb"}]}]`,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifyGroup := &NotifyGroup{
				BaseModel:   tt.fields.BaseModel,
				Name:        tt.fields.Name,
				ScopeType:   tt.fields.ScopeType,
				ScopeID:     tt.fields.ScopeID,
				OrgID:       tt.fields.OrgID,
				TargetData:  tt.fields.TargetData,
				Label:       tt.fields.Label,
				ClusterName: tt.fields.ClusterName,
				AutoCreate:  tt.fields.AutoCreate,
				Creator:     tt.fields.Creator,
			}
			if got := notifyGroup.ToApiData(); got != nil {
				fmt.Printf("ToApiData() = %+v", got)
			}
		})
	}
}
