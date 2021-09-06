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
	bt := `{"id":17046,"pipelineID":18564,"type":"auto-test-plan","meta":{"data":"{\"domain\":\"https://openapi.hkci.terminus.io\",\"header\":{\"Cookie\":\"u_c_captain_hkci_local=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJsb2dpbiIsInBhdGgiOiIvIiwidG9rZW5LZXkiOiI5NTA4NjI2ZTczZDdiODdjYjY1N2M3YTUzMGY1NWM4ZDg0OTJjYjc0YjFmZDdlZTNlNjhmZThkNThmZjg4ZDM2IiwibmJmIjoxNjI5ODkwMTMyLCJkb21haW4iOiJ0ZXJtaW51cy5pbyIsImlzcyI6ImRyYWNvIiwidGVuYW50SWQiOjEsImV4cGlyZV90aW1lIjo2MDQ4MDAsImV4cCI6MTYzMDQ5NDkzMiwiaWF0IjoxNjI5ODkwMTMyfQ.jTI2D54sC-aM90p8TX9pfeXXwOFi8Yns6smQhNSVZKA; OPENAPISESSION=6cccd9ac-38f0-4906-9da8-0143186dd631; OPENAPI-CSRF-TOKEN=12bde30a9db3d9aad0159ecbd9ae79b89005c3a5eca326a2b7d0f349cf897441342e1bc368ab501084ed6de62d8e57ca\",\"cluster-id\":\"2\",\"cluster-name\":\"ZXJkYS1ob25na29uZw==\",\"org\":\"erda\",\"project-id\":\"13\"},\"global\":{\"111\":{\"name\":\"111\",\"type\":\"string\",\"value\":\"111\",\"desc\":\"111\"}}}"},"creatorID":"","updaterID":"","createdAt":"2021-09-03T17:25:48+08:00","updatedAt":"2021-09-03T17:25:48+08:00"}`
	err := json.Unmarshal([]byte(bt), &m)
	assert.NoError(t, err)
	c, err := convertReportToConfig(m)
	assert.NoError(t, err)
	want := apistructs.AutoTestAPIConfig{
		Domain: "https://openapi.hkci.terminus.io",
		Header: map[string]string{
			"Cookie":       "u_c_captain_hkci_local=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJsb2dpbiIsInBhdGgiOiIvIiwidG9rZW5LZXkiOiI5NTA4NjI2ZTczZDdiODdjYjY1N2M3YTUzMGY1NWM4ZDg0OTJjYjc0YjFmZDdlZTNlNjhmZThkNThmZjg4ZDM2IiwibmJmIjoxNjI5ODkwMTMyLCJkb21haW4iOiJ0ZXJtaW51cy5pbyIsImlzcyI6ImRyYWNvIiwidGVuYW50SWQiOjEsImV4cGlyZV90aW1lIjo2MDQ4MDAsImV4cCI6MTYzMDQ5NDkzMiwiaWF0IjoxNjI5ODkwMTMyfQ.jTI2D54sC-aM90p8TX9pfeXXwOFi8Yns6smQhNSVZKA; OPENAPISESSION=6cccd9ac-38f0-4906-9da8-0143186dd631; OPENAPI-CSRF-TOKEN=12bde30a9db3d9aad0159ecbd9ae79b89005c3a5eca326a2b7d0f349cf897441342e1bc368ab501084ed6de62d8e57ca",
			"cluster-id":   "2",
			"cluster-name": "ZXJkYS1ob25na29uZw==",
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
