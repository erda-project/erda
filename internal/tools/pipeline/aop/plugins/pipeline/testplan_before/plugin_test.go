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

package testplan_before

import (
	"testing"

	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/services/autotest"
	"github.com/erda-project/erda/internal/tools/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/report"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

func Test_checkPipelineYmlName(t *testing.T) {
	isTestPlan := checkPipelineYmlName("autotest-scene-12")
	assert.False(t, isTestPlan)

	isTestPlan = checkPipelineYmlName("autotest-plan-12")
	assert.True(t, isTestPlan)
}

func TestHandle(t *testing.T) {
	p := &provider{}
	ctx := &aoptypes.TuneContext{
		SDK: aoptypes.SDK{
			Pipeline: spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					PipelineSource:  apistructs.PipelineSourceAutoTest,
					PipelineYmlName: "autotest-plan-1",
				},
				PipelineExtra: spec.PipelineExtra{
					Snapshot: spec.Snapshot{
						Secrets: map[string]string{
							autotest.CmsCfgKeyAPIGlobalConfig: "config",
						},
					},
				},
			},
			Report: &report.MockReport{},
		},
	}
	err := p.Handle(ctx)
	assert.NoError(t, err)
}
