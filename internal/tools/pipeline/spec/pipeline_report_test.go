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

package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestConvertToPB(t *testing.T) {
	report := &PipelineReport{
		ID:         1,
		PipelineID: 1,
		Type:       apistructs.PipelineReportTypeAPITest,
		Meta: apistructs.PipelineReportMeta{
			"name": "test",
		},
	}
	pbReport, err := report.ConvertToPB()
	assert.NoError(t, err)
	assert.Equal(t, "test", pbReport.Meta.AsMap()["name"])
	assert.Equal(t, string(apistructs.PipelineReportTypeAPITest), pbReport.Type)
}
