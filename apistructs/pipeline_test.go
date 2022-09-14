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

package apistructs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
)

func TestToPipelineRunParamsPB(t *testing.T) {
	params := PipelineRunParamsWithValue{
		{
			PipelineRunParam: PipelineRunParam{
				Name:  "param1",
				Value: "value1",
			},
		},
		{
			PipelineRunParam: PipelineRunParam{
				Name:  "param2",
				Value: "value2",
			},
		},
	}
	pbParams, err := params.ToPipelineRunParamsPB()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pbParams))
}

func TestIsPipelineDefinitionReqEmpty(t *testing.T) {
	type arg struct {
		definition *pipelinepb.PipelineDefinitionRequest
	}
	tests := []struct {
		name string
		arg  arg
		want bool
	}{
		{
			name: "nil definition",
			arg: arg{
				definition: nil,
			},
			want: true,
		},
		{
			name: "empty string",
			arg: arg{
				definition: &pipelinepb.PipelineDefinitionRequest{},
			},
			want: true,
		},
		{
			name: "has name",
			arg: arg{
				definition: &pipelinepb.PipelineDefinitionRequest{
					Name: "name",
				},
			},
		},
		{
			name: "has creators",
			arg: arg{
				definition: &pipelinepb.PipelineDefinitionRequest{
					Creators: []string{"user1"},
				},
			},
			want: false,
		},
		{
			name: "has source",
			arg: arg{
				definition: &pipelinepb.PipelineDefinitionRequest{
					SourceRemotes: []string{
						"git",
					},
				},
			},
			want: false,
		},
		{
			name: "has location",
			arg: arg{
				definition: &pipelinepb.PipelineDefinitionRequest{
					Location: "dice",
				},
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isEmpty := IsPipelineDefinitionReqEmpty(tc.arg.definition)
			assert.Equal(t, tc.want, isEmpty)
		})
	}
}

func TestPostHandlePBQueryString(t *testing.T) {
	now := time.Now()
	timeFormat := "2006-01-02T15:04:05"
	req := &pipelinepb.PipelinePagingRequest{
		Branches:        "master,develop",
		Sources:         "dice,cdp",
		Statuses:        "Running,Queue",
		YmlNames:        "yml1,yml2",
		MustMatchLabels: `{"FDP_PROJECT_ID":"1","orgID":"1000028"}`,
		AnyMatchLabels:  `{"FDP_PROJECT_ID":"1","orgID":"1000028"}`,
		MustMatchLabel: []string{
			"k1=v1",
			"k2=v2",
			"k3=v3",
		},
		AnyMatchLabel: []string{
			"k1=v1",
			"k2=v2",
		},
		StartTimeBeginTimestamp:   now.Unix(),
		EndTimeBeginTimestamp:     now.Unix(),
		StartedAt:                 now.Format(timeFormat),
		EndedAt:                   now.Format(timeFormat),
		StartTimeCreatedTimestamp: now.Unix(),
		EndTimeCreatedTimestamp:   now.Unix(),
	}
	err := PostHandlePBQueryString(req)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(req.MustMatchLabelsJSON))
	assert.Equal(t, 4, len(req.AnyMatchLabelsJSON))
}
