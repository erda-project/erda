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

package pipelineTable

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func TestDecodeToCustomInParams(t *testing.T) {
	pipelineTable := &PipelineTable{
		InParams: &InParams{
			FrontendProjectID: "123",
			FrontendAppID:     "456",
		},
	}

	stdInParams := &cptype.ExtraMap{}
	pipelineTable.DecodeToCustomInParams(stdInParams, nil)

	assert.Equal(t, uint64(123), pipelineTable.InParams.ProjectID)
	assert.Equal(t, uint64(456), pipelineTable.InParams.AppID)
}
