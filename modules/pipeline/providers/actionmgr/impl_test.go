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

package actionmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestMakeActionTypeVersion(t *testing.T) {
	p := &provider{}
	item := p.MakeActionTypeVersion(&pipelineyml.Action{Type: "git", Version: "1.0"})
	assert.Equal(t, item, "git@1.0")

	item = p.MakeActionTypeVersion(&pipelineyml.Action{Type: "git"})
	assert.Equal(t, item, "git")
}

func Test_provider_MakeActionLocationsBySource(t *testing.T) {
	p := &provider{}

	type args struct {
		inputSource               apistructs.PipelineSource
		expectedOutputLocationNum int
		expectedOutputLocations   []string
	}

	cases := []args{
		// fdp
		{
			inputSource:               apistructs.PipelineSourceCDPDev,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeFDP.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceCDPTest,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeFDP.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceCDPStaging,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeFDP.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceCDPProd,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeFDP.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceBigData,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeFDP.String() + "/"},
		},
		// cicd
		{
			inputSource:               apistructs.PipelineSourceDice,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeCICD.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceProject,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeCICD.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceProjectLocal,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeCICD.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceOps,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeCICD.String() + "/"},
		},
		{
			inputSource:               apistructs.PipelineSourceQA,
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeCICD.String() + "/"},
		},
		// default
		{
			inputSource:               "unknown",
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeDefault.String() + "/"},
		},
		{
			inputSource:               "",
			expectedOutputLocationNum: 1,
			expectedOutputLocations:   []string{apistructs.PipelineTypeDefault.String() + "/"},
		},
	}

	for _, c := range cases {
		locations := p.MakeActionLocationsBySource(c.inputSource)
		if len(locations) != c.expectedOutputLocationNum {
			t.Fatalf("location num mismatch, actual: %d, expected: %d", len(locations), c.expectedOutputLocationNum)
		}
		for _, el := range c.expectedOutputLocations {
			if !strutil.Exist(c.expectedOutputLocations, el) {
				t.Fatalf("missing expected output location %s", el)
			}
		}
	}
}
