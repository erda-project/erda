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

package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestGetComponentName(t *testing.T) {
	assert.Equal(t, "addon-nexus", strings.Split("addon-nexus.default:8081", ".")[0])
}

func TestLoad(t *testing.T) {
	baseDir, err := os.MkdirTemp(os.TempDir(), "erda-configs")
	assert.NoErrorf(t, err, "failed to create temp dir")
	defer os.RemoveAll(baseDir)
	err = os.Mkdir(filepath.Join(baseDir, "permission"), 0755)
	assert.NoErrorf(t, err, "failed to create dir")
	os.Mkdir(filepath.Join(baseDir, "audit"), 0755)
	f, err := os.Create(filepath.Join(baseDir, "audit/template.json"))
	assert.NoErrorf(t, err, "failed to write file")
	f.WriteString(`{
  "cancelPipeline": {
    "desc": "取消流水线",
    "success": {
      "zh": "在应用 [@projectName](project) / [@appName](app) 中，取消执行 [流水线](pipeline)",
      "en": "In the application [@projectName](project) / [@appName](app), cancel the execution of [pipeline](pipeline)"
    },
    "fail": {}
  }}`)
	f.Close()
	os.Setenv("ERDA_CONFIGS_BASE_PATH", baseDir)
	Load()
	assert.Equal(t, uint64(365), OrgAuditDefaultRetentionDays())
}
