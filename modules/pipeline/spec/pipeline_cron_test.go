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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestGenCompensateCreatePipelineReqNormalLabels(t *testing.T) {
	pc := PipelineCron{ID: 1, Extra: PipelineCronExtra{
		NormalLabels: map[string]string{
			"org":                               "erda",
			apistructs.LabelPipelineTriggerMode: "dice",
		},
	}}
	now := time.Now()
	normalLabels := pc.GenCompensateCreatePipelineReqNormalLabels(now)
	assert.Equal(t, "erda", normalLabels["org"])
	assert.Equal(t, apistructs.PipelineTriggerModeCron.String(), normalLabels[apistructs.LabelPipelineTriggerMode])
	assert.Equal(t, strconv.FormatInt(now.UnixNano(), 10), normalLabels[apistructs.LabelPipelineCronTriggerTime])
}

func TestGenCompensateCreatePipelineReqFilterLabels(t *testing.T) {
	pc := PipelineCron{ID: 1, Extra: PipelineCronExtra{
		FilterLabels: map[string]string{
			"org":                               "erda",
			apistructs.LabelPipelineTriggerMode: "dice",
		},
	}}
	filterLabels := pc.GenCompensateCreatePipelineReqFilterLabels()
	assert.Equal(t, "true", filterLabels[apistructs.LabelPipelineCronCompensated])
	assert.Equal(t, apistructs.PipelineTriggerModeCron.String(), filterLabels[apistructs.LabelPipelineTriggerMode])
}
