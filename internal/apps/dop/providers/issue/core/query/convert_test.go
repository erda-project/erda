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

package query

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/i18n"
)

func Test_provider_getIssueExportDataI18n(t *testing.T) {
	bdl := bundle.New(bundle.WithI18nLoader(&i18n.LocaleResourceLoader{}))
	m := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetLocale",
		func(bdl *bundle.Bundle, local ...string) *i18n.LocaleResource {
			return &i18n.LocaleResource{}
		})
	defer m.Unpatch()

	p := &provider{bdl: bdl}
	content := ",a,b,c"
	expected := []string{"", "a", "b", "c"}
	strs := p.getIssueExportDataI18n("testKey", content)
	assert.Equal(t, strs, expected)
}

func Test_provider_ConvertWithoutButton(t *testing.T) {
	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetLabelRelationsByRef",
		func(d *dao.DBClient, refType string, refID string) ([]dao.LabelRelation, error) {
			return []dao.LabelRelation{
				{
					BaseModel: dbengine.BaseModel{
						ID: 1,
					},
				},
			}, nil
		},
	)
	defer p1.Unpatch()

	m := dao.Issue{
		BaseModel: dbengine.BaseModel{ID: 1},
	}
	p := &provider{db: db}
	pl := map[uint64]apistructs.ProjectLabel{
		1: {
			ID: 1,
		},
	}
	got := p.ConvertWithoutButton(m, true, []uint64{1}, false, pl)
	assert.Equal(t, m.ID, uint64(got.Id))
}

func Test_provider_Convert(t *testing.T) {
	var db *dao.DBClient

	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssueSubscribersByIssueID",
		func(d *dao.DBClient, issueID int64) ([]dao.IssueSubscriber, error) {
			return []dao.IssueSubscriber{
				{
					IssueID: 1,
					UserID:  "2",
				},
			}, nil
		},
	)
	defer p1.Unpatch()

	p := &provider{db: db}
	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertWithoutButton",
		func(p *provider, model dao.Issue,
			needQueryLabelRef bool, labelIDs []uint64,
			needQueryLabels bool, projectLabels map[uint64]apistructs.ProjectLabel,
		) *pb.Issue {
			return &pb.Issue{
				Id: 1,
			}
		},
	)
	defer p2.Unpatch()

	p3 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "GenerateButton",
		func(p *provider, issueModel dao.Issue, identityInfo *commonpb.IdentityInfo,
			permCheckItems map[string]bool, store map[string]bool, relations map[dao.IssueStateRelation]bool,
			typeState map[string][]*pb.IssueStateButton) ([]*pb.IssueStateButton, error) {
			return []*pb.IssueStateButton{
				{StateID: 1},
			}, nil
		},
	)
	defer p3.Unpatch()

	p4 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "GetIssuePropertyInstance",
		func(p *provider, req *pb.GetIssuePropertyInstanceRequest) (*pb.IssueAndPropertyAndValue, error) {
			return &pb.IssueAndPropertyAndValue{
				IssueID: req.IssueID,
				Property: []*pb.IssuePropertyExtraProperty{
					{
						PropertyID: 1,
					},
				},
			}, nil
		})
	defer p4.Unpatch()

	_, err := p.Convert(dao.Issue{BaseModel: dbengine.BaseModel{ID: 1}}, &commonpb.IdentityInfo{
		UserID: "1",
		OrgID:  "1",
	})

	assert.NoError(t, err)
}

func Test_provider_BatchConvert(t *testing.T) {
	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "BatchQueryIssueLabelIDMap",
		func(d *dao.DBClient, issueIDs []int64) (map[uint64][]uint64, error) {
			return map[uint64][]uint64{
				1: {1, 2},
			}, nil
		},
	)
	defer p1.Unpatch()

	bdl := bundle.New()
	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "ListLabel",
		func(bdl *bundle.Bundle, req apistructs.ProjectLabelListRequest) (*apistructs.ProjectLabelListResponseData, error) {
			return &apistructs.ProjectLabelListResponseData{
				Total: 2,
				List: []apistructs.ProjectLabel{
					{
						ID: 1,
					},
				},
			}, nil
		})
	defer p2.Unpatch()

	p := &provider{db: db, bdl: bdl}
	p3 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertWithoutButton",
		func(p *provider, model dao.Issue,
			needQueryLabelRef bool, labelIDs []uint64,
			needQueryLabels bool, projectLabels map[uint64]apistructs.ProjectLabel,
		) *pb.Issue {
			return &pb.Issue{
				Id: 1,
			}
		},
	)
	defer p3.Unpatch()

	p4 := monkey.PatchInstanceMethod(reflect.TypeOf(p), "GenerateButtonMap",
		func(p *provider, projectID uint64, issueTypes []string) (map[string]map[int64][]*pb.IssueStateButton, error) {
			return nil, nil
		},
	)
	defer p4.Unpatch()
	_, err := p.BatchConvert([]dao.Issue{{ProjectID: 1}}, []string{"TASK", "BUG"})
	assert.NoError(t, err)
}

func Test_rangePropertyMap(t *testing.T) {
	propertyMap := make(map[int64][]dao.IssuePropertyRelation)
	properties := []dao.IssuePropertyRelation{{PropertyID: 1, IssueID: 1}, {PropertyID: 2, IssueID: 2}}
	for _, v := range properties {
		propertyMap[v.IssueID] = append(propertyMap[v.IssueID], v)
	}
	assert.Equal(t, 2, len(propertyMap))
	assert.Equal(t, int64(1), propertyMap[1][0].IssueID)
	assert.Equal(t, int64(1), propertyMap[1][0].PropertyID)
	assert.Equal(t, int64(2), propertyMap[2][0].IssueID)
	assert.Equal(t, int64(2), propertyMap[2][0].PropertyID)
}

func Test_getCustomPropertyColumnValue(t *testing.T) {
	type args struct {
		pro       *pb.IssuePropertyIndex
		relations []dao.IssuePropertyRelation
		mp        map[PropertyEnumPair]string
		users     map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no property, no panic",
			args: args{pro: nil},
			want: "",
		},
		{
			name: "no relation, no panic",
			args: args{pro: &pb.IssuePropertyIndex{}},
			want: "",
		},
		{
			name: "no mp, no panic",
			args: args{
				pro:       &pb.IssuePropertyIndex{PropertyID: 1},
				relations: []dao.IssuePropertyRelation{{PropertyID: 1}},
				mp:        nil,
			},
			want: "",
		},
		{
			name: "arbitrary value",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Text,
				},
				relations: []dao.IssuePropertyRelation{{PropertyID: 1, ArbitraryValue: "text"}},
				mp:        nil,
			},
			want: "text",
		},
		{
			name: "arbitrary value, but not match",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Text,
				},
				relations: []dao.IssuePropertyRelation{{PropertyID: 2, ArbitraryValue: "text"}},
				mp:        nil,
			},
			want: "",
		},
		{
			name: "select",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Select,
				},
				relations: []dao.IssuePropertyRelation{{PropertyID: 1, PropertyValueID: 1}},
				mp: map[PropertyEnumPair]string{
					{
						PropertyID: 1,
						ValueID:    1,
					}: "select value",
				},
			},
			want: "select value",
		},
		{
			name: "select, but not match",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Select,
				},
				relations: []dao.IssuePropertyRelation{{PropertyID: 1, PropertyValueID: 1}},
				mp: map[PropertyEnumPair]string{
					{PropertyID: 1, ValueID: 2}: "select value",
				},
			},
			want: "",
		},
		{
			name: "multi select",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_MultiSelect,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 1, PropertyValueID: 1},
					{PropertyID: 1, PropertyValueID: 2},
				},
				mp: map[PropertyEnumPair]string{
					{PropertyID: 1, ValueID: 1}: "m11",
					{PropertyID: 1, ValueID: 2}: "m12",
				},
			},
			want: "m11,m12",
		},
		{
			name: "multi select, but no match",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_MultiSelect,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 2, PropertyValueID: 1},
					{PropertyID: 2, PropertyValueID: 2},
				},
				mp: map[PropertyEnumPair]string{
					{PropertyID: 1, ValueID: 1}: "m11",
					{PropertyID: 1, ValueID: 2}: "m12",
				},
			},
			want: "",
		},
		{
			name: "check box",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_CheckBox,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 1, PropertyValueID: 1},
					{PropertyID: 1, PropertyValueID: 2},
				},
				mp: map[PropertyEnumPair]string{
					{PropertyID: 1, ValueID: 1}: "m11",
					{PropertyID: 1, ValueID: 2}: "m12",
				},
			},
			want: "m11,m12",
		},
		{
			name: "check box, but no match",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_CheckBox,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 2, PropertyValueID: 1},
					{PropertyID: 2, PropertyValueID: 2},
				},
				mp: map[PropertyEnumPair]string{
					{PropertyID: 1, ValueID: 1}: "m11",
					{PropertyID: 1, ValueID: 2}: "m12",
				},
			},
			want: "",
		},
		{
			name: "check date",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Date,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 1, ArbitraryValue: "2023-01-17T00:00:00+08:00"},
				},
				mp:    map[PropertyEnumPair]string{},
				users: map[string]string{},
			},
			want: "2023-01-17 00:00:00",
		},
		{
			name: "check person",
			args: args{
				pro: &pb.IssuePropertyIndex{
					PropertyID:   1,
					PropertyType: pb.PropertyTypeEnum_Person,
				},
				relations: []dao.IssuePropertyRelation{
					{PropertyID: 1, ArbitraryValue: "1001"},
				},
				mp: map[PropertyEnumPair]string{},
				users: map[string]string{
					"1001": "nickname",
				},
			},
			want: "nickname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberMap := make(map[string]apistructs.Member)
			for k, v := range tt.args.users {
				memberMap[k] = apistructs.Member{
					UserID: k,
					Nick:   v,
				}
			}
			if got := GetCustomPropertyColumnValue(tt.args.pro, tt.args.relations, tt.args.mp, memberMap); got != tt.want {
				t.Errorf("GetCustomPropertyColumnValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
