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

package excel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/i18n"
)

func Test_generateExcelTitleByRowStruct(t *testing.T) {
	rows := generateExcelTitleByRowStruct(i18n.LocaleResource{}, []*pb.IssuePropertyIndex{
		{
			PropertyName:      "需求自定义字段-1",
			PropertyIssueType: pb.PropertyIssueTypeEnum_REQUIREMENT,
			Index:             0,
		},
		{
			PropertyName:      "任务自定义字段-1",
			PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
			Index:             1,
		},
		{
			PropertyName:      "任务自定义字段-2",
			PropertyIssueType: pb.PropertyIssueTypeEnum_TASK,
			Index:             2,
		},
		{
			PropertyName:      "缺陷自定义字段-1",
			PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
			Index:             3,
		},
		{
			PropertyName:      "缺陷自定义字段-2",
			PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
			Index:             4,
		},
		{
			PropertyName:      "缺陷自定义字段-3",
			PropertyIssueType: pb.PropertyIssueTypeEnum_BUG,
			Index:             5,
		},
	})
	f, err := os.OpenFile("./gen.xlsx", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	assert.NoError(t, err)
	defer f.Close()
	err = excel.ExportExcelByCell(f, rows, "issues")
	assert.NoError(t, err)
}
