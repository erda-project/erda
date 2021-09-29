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

package issue

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/magiconair/properties/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issuerelated"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/i18n"
)

func TestIssue_getIssueExportDataI18n(t *testing.T) {
	bdl := bundle.New(bundle.WithI18nLoader(&i18n.LocaleResourceLoader{}))
	m := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetLocale",
		func(bdl *bundle.Bundle, local ...string) *i18n.LocaleResource {
			return &i18n.LocaleResource{}
		})
	defer m.Unpatch()

	svc := New(WithBundle(bdl))
	content := ",a,b,c"
	expected := []string{"", "a", "b", "c"}
	strs := svc.getIssueExportDataI18n("testKey", content)
	assert.Equal(t, strs, expected)
}

func TestDecodeFromExcelFile(t *testing.T) {
	tm := monkey.Patch(excel.Decode, func(r io.Reader) ([][][]string, error) {
		return [][][]string{
			[][]string{
				[]string{"ID", "标题", "内容", "状态", "创建人", "处理人", "负责人", "任务类型或缺陷引入源", "优先级", "所属迭代", "复杂度", "严重程度", "标签", "类型", "截止时间", "创建时间", "被以下事项关联", "预估时间"},
				[]string{"1", "erda", "erda", "待处理", "erda", "erda", "erda", "缺陷", "低", "1.1", "中", "一般", "", "缺陷", "", "2021-09-26 15:19:00", "", "", "", ""},
				[]string{"a", "erda", "erda", "待处理", "erda", "erda", "erda", "缺陷", "低", "1.1", "中", "一般", "", "缺陷", "", "2021-09-26 15:19:00", "", "", "", ""},
				[]string{"2", "erda", "erda", "待处理", "erda", "erda", "erda", "缺陷", "低", "1.1", "中", "一般", "", "缺陷", "", "2021-09-26 15:19:00", "a,b", "", "", ""},
				[]string{"2", "erda", "erda", "待处理", "erda", "erda", "erda", "缺陷", "低", "1.1", "中", "一般", "", "缺陷", "", "2021-09-26 15:19:00", "", "", "", ""},
			},
		}, nil
	})
	defer tm.Unpatch()

	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssuesStatesByProjectID",
		func(d *dao.DBClient, projectID uint64, issueType apistructs.IssueType) ([]dao.IssueState, error) {
			return []dao.IssueState{
				{
					ProjectID: projectID,
					Name:      "待处理",
				},
			}, nil
		},
	)
	defer p1.Unpatch()

	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "FindIterations",
		func(d *dao.DBClient, projectID uint64) (iterations []dao.Iteration, err error) {
			return []dao.Iteration{
				{
					ProjectID: projectID,
					Title:     "1.1",
				},
			}, nil
		},
	)
	defer p2.Unpatch()

	svc := &Issue{db: db}
	_, _, _, _, _, _, err := svc.decodeFromExcelFile(apistructs.IssueImportExcelRequest{ProjectID: 1, OrgID: 1}, strings.NewReader(""), []apistructs.IssuePropertyIndex{})
	assert.Equal(t, nil, err)
}

func TestConvertWithoutButton(t *testing.T) {
	svc := &Issue{}
	i := svc.ConvertWithoutButton(dao.Issue{ManHour: "{\"estimateTime\": 3}"}, false, []uint64{}, false, map[uint64]apistructs.ProjectLabel{})
	assert.Equal(t, i.ManHour.EstimateTime, int64(3))
}

func TestConvertIssueToExcelList(t *testing.T) {
	var db *dao.DBClient
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetIssuesStatesByProjectID",
		func(d *dao.DBClient, projectID uint64, issueType apistructs.IssueType) ([]dao.IssueState, error) {
			return []dao.IssueState{
				{
					ProjectID: projectID,
					Name:      "待处理",
				},
			}, nil
		},
	)
	defer p1.Unpatch()

	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "PagingIterations",
		func(d *dao.DBClient, req apistructs.IterationPagingRequest) ([]dao.Iteration, uint64, error) {
			return []dao.Iteration{
				{
					ProjectID: req.ProjectID,
					Title:     "1.1",
				},
			}, 1, nil
		},
	)
	defer p2.Unpatch()

	bdl := bundle.New(bundle.WithI18nLoader(&i18n.LocaleResourceLoader{}))
	p3 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetLocale",
		func(bdl *bundle.Bundle, local ...string) *i18n.LocaleResource {
			return &i18n.LocaleResource{}
		})
	defer p3.Unpatch()

	p4 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "PagingPropertyRelationByIDs",
		func(d *dao.DBClient, issueID []int64) ([]dao.IssuePropertyRelation, error) {
			return []dao.IssuePropertyRelation{
				{
					IssueID: 1,
				},
			}, nil
		},
	)
	defer p4.Unpatch()

	related := &issuerelated.IssueRelated{}
	p5 := monkey.PatchInstanceMethod(reflect.TypeOf(related), "GetIssueRelationsByIssueIDs",
		func(i *issuerelated.IssueRelated, issueID uint64) ([]uint64, []uint64, error) {
			return []uint64{}, []uint64{}, nil
		},
	)
	defer p5.Unpatch()

	svc := New(WithDBClient(db), WithIssueRelated(related), WithBundle(bdl))
	_, err := svc.convertIssueToExcelList([]apistructs.Issue{}, []apistructs.IssuePropertyIndex{}, 1, false, map[issueStage]string{}, "cn")
	assert.Equal(t, err, nil)
}
