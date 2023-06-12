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

package issueexcel

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/dao"
)

func TestNewIssueSheetColumnUUID(t *testing.T) {
	uuid := NewIssueSheetColumnUUID("Common", "ID")
	assert.Equal(t, uuid.String(), strings.Join([]string{"Common", "ID", "ID"}, issueSheetColumnUUIDSplitter))
	uuid = NewIssueSheetColumnUUID("ID")
	assert.Equal(t, uuid.String(), strings.Join([]string{"ID", "ID", "ID"}, issueSheetColumnUUIDSplitter))
	uuid = NewIssueSheetColumnUUID("")
	assert.Panicsf(t, func() { uuid.String() }, "uuid should not be empty")
	uuid.AddPart("Common")
	uuid.AddPart("ID")
	assert.Equal(t, uuid.String(), strings.Join([]string{"Common", "ID", "ID"}, issueSheetColumnUUIDSplitter))
}

func Test_genIssueSheetTitleAndDataByColumn(t *testing.T) {
	data := DataForFulfill{
		ExportOnly: DataForFulfillExportOnly{
			Issues: []*pb.Issue{
				{
					Id:              1,
					RelatedIssueIDs: []uint64{2, 3},
				},
			},
			ConnectionMap: map[int64][]int64{
				1: {2, 3},
			},
		},
		CustomFieldMapByTypeName: map[pb.PropertyIssueTypeEnum_PropertyIssueType]map[string]*pb.IssuePropertyIndex{},
		StateMap:                 map[int64]string{},
		StageMap:                 map[query.IssueStage]string{},
		IterationMapByID: map[int64]*dao.Iteration{
			0: {},
		},
	}
	info, err := data.genIssueSheetTitleAndDataByColumn()
	assert.NoError(t, err)
	fmt.Printf("%+v\n", info)
}

func Test_getStringCellValue(t *testing.T) {
	common := IssueSheetModelCommon{
		ID:                 1,
		ConnectionIssueIDs: []int64{2, 3},
	}
	commonValue := reflect.ValueOf(&common).Elem()
	valueField := commonValue.FieldByName("ConnectionIssueIDs")
	typeField, ok := commonValue.Type().FieldByName("ConnectionIssueIDs")
	assert.True(t, ok)
	assert.Equal(t, getStringCellValue(typeField, valueField), "2,3")

	common = IssueSheetModelCommon{
		ID:                 0,
		ConnectionIssueIDs: []int64{2, -3},
	}
	commonValue = reflect.ValueOf(&common).Elem()
	valueField = commonValue.FieldByName("ConnectionIssueIDs")
	typeField, ok = commonValue.Type().FieldByName("ConnectionIssueIDs")
	assert.True(t, ok)
	assert.Equal(t, getStringCellValue(typeField, valueField), "2,L3")
	valueField = commonValue.FieldByName("ID")
	typeField, ok = commonValue.Type().FieldByName("ID")
	assert.True(t, ok)
	assert.Equal(t, getStringCellValue(typeField, valueField), "")
}
