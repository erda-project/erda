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

package cmd

import "testing"

func TestNormalizePipelineYmlName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain root pipeline", in: "pipeline.yml", want: "pipeline.yml"},
		{name: "dot slash root pipeline", in: "./pipeline.yml", want: "pipeline.yml"},
		{name: "nested pipeline", in: ".erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
		{name: "dot slash nested pipeline", in: "./.erda/pipelines/java-demo.yml", want: ".erda/pipelines/java-demo.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePipelineYmlName(tt.in); got != tt.want {
				t.Fatalf("normalizePipelineYmlName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
