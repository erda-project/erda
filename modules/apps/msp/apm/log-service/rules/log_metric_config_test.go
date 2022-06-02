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

package rules

import (
	"testing"

	"gotest.tools/assert"

	"github.com/erda-project/erda/modules/apps/msp/apm/log-service/rules/db"
)

func Test_convertToMetricMeta(t *testing.T) {
	p := &provider{}
	_, err := p.convertToMetricMeta(&db.LogMetricConfig{
		Metric: "log_test",
		Name:   "log_test",
	})

	assert.NilError(t, err)
}
