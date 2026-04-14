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

package common

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestLatestPipelineID(t *testing.T) {
	tests := []struct {
		name      string
		pipelines []apistructs.PagePipeline
		wantID    uint64
		wantOK    bool
	}{
		{
			name:      "empty pipelines",
			pipelines: nil,
			wantID:    0,
			wantOK:    false,
		},
		{
			name: "pick max id",
			pipelines: []apistructs.PagePipeline{
				{ID: 1002},
				{ID: 998},
				{ID: 1200},
			},
			wantID: 1200,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := latestPipelineID(tt.pipelines)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("latestPipelineID() = (%d, %v), want (%d, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}
