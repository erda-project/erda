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

package executeInfo

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_convertReportToConfig(t *testing.T) {
	m := apistructs.PipelineReport{}
	bt := `{"id":123,"pipelineID":123,"type":"auto-test-plan","meta":{"data":"{\"domain\":\"domain\",\"header\":{\"Cookie\":\"cookie\",\"cluster-id\":\"2\",\"cluster-name\":\"name\",\"org\":\"erda\",\"project-id\":\"13\"},\"global\":{\"111\":{\"name\":\"111\",\"type\":\"string\",\"value\":\"111\",\"desc\":\"111\"}}}"},"creatorID":"","updaterID":"","createdAt":"2021-09-03T17:25:48+08:00","updatedAt":"2021-09-03T17:25:48+08:00"}`
	err := json.Unmarshal([]byte(bt), &m)
	assert.NoError(t, err)
	c, err := convertReportToConfig(m)
	assert.NoError(t, err)
	want := apistructs.AutoTestAPIConfig{
		Domain: "domain",
		Header: map[string]string{
			"Cookie":       "cookie",
			"cluster-id":   "2",
			"cluster-name": "name",
			"org":          "erda",
			"project-id":   "13",
		},
		Global: map[string]apistructs.AutoTestConfigItem{
			"111": {
				Name:  "111",
				Type:  "string",
				Value: "111",
				Desc:  "111",
			},
		}}
	assert.Equal(t, want, c)
}
