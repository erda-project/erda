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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBranch(t *testing.T) {
	ss := []struct {
		ref  string
		Want string
	}{
		{"", ""},
		{"refs/heads/", ""},
		{"refs/heads/master", "master"},
		{"refs/heads/feature/test", "feature/test"},
	}
	for _, v := range ss {
		assert.Equal(t, v.Want, getBranch(v.ref))
	}
}

func TestIsPipelineYmlPath(t *testing.T) {
	ss := []struct {
		path string
		want bool
	}{
		{"pipeline.yml", true},
		{".dice/pipelines/a.yml", true},
		{"", false},
		{"dice/pipeline.yml", false},
	}
	for _, v := range ss {
		assert.Equal(t, v.want, isPipelineYmlPath(v.path))
	}

}
