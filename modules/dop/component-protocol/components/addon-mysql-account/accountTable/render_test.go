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

package accountTable

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/accountTable/table"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
)

func Test_comp_getDatum(t *testing.T) {
	type fields struct {
		ac      *common.AccountData
		pg      *common.PageDataAccount
		userIDs []string
	}
	now := time.Now()
	type args struct {
		item *addonmysqlpb.MySQLAccount
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]table.ColumnData
	}{
		{
			name: "t1",
			fields: fields{
				ac: &common.AccountData{
					Attachments:     nil,
					AttachmentMap:   nil,
					Accounts:        nil,
					AccountMap:      nil,
					AccountRefCount: map[string]int{},
					Apps:            nil,
					AppMap:          nil,
				},
				pg: &common.PageDataAccount{
					ProjectID:             24,
					InstanceID:            "123",
					AccountID:             "555",
					ShowDeleteModal:       false,
					ShowViewPasswordModal: false,
					FilterValues:          nil,
				},
				userIDs: nil,
			},
			args: args{
				item: &addonmysqlpb.MySQLAccount{
					Id:         "111",
					InstanceId: "123",
					Creator:    "333",
					CreateAt:   timestamppb.New(now),
					Username:   "user",
					Password:   "pass",
				},
			},
			want: map[string]table.ColumnData{
				"attachments": {
					Value:      "未被使用",
					RenderType: "text",
				},
				"createdAt": {
					Value:      timestamppb.New(now).AsTime().Format(time.RFC3339),
					RenderType: "datePicker",
				},
				"creator": {
					Value:      "333",
					RenderType: "userAvatar",
				},
				"username": {
					Value:      "user",
					RenderType: "text",
				},
				"operate": {
					Value:      "",
					RenderType: "tableOperation",
					Tags:       nil,
					Operations: map[string]*table.Operation{
						"viewPassword": {
							Key:    "viewPassword",
							Text:   "查看密码",
							Reload: true,
							Meta: map[string]string{
								"id": "111",
							},
							Disabled:    true,
							DisabledTip: "您没有权限查看密码，请联系项目管理员",
							ShowIndex:   1,
						},
						"delete": {
							Key:    "delete",
							Text:   "删除",
							Reload: true,
							Meta: map[string]string{
								"id": "111",
							},
							Disabled:    true,
							DisabledTip: "您没有权限删除账号，请联系项目管理员",
							ShowIndex:   2,
							Confirm:     "是否确认删除",
							SuccessMsg:  "删除成功",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &comp{
				ac:      tt.fields.ac,
				pg:      tt.fields.pg,
				userIDs: tt.fields.userIDs,
			}
			if got := f.getDatum(tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDatum() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_getTitles(t *testing.T) {
	tests := []struct {
		name string
		want []*table.ColumnTitle
	}{
		{name: "t1", want: getTitles()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTitles(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTitles() = %v, want %v", got, tt.want)
			}
		})
	}
}
