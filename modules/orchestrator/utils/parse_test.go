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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeploymentOrderShowName(t *testing.T) {
	type args struct {
		orderName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pipeline",
			args: args{
				orderName: "master",
			},
			want: "master",
		},
		{
			name: "project-error",
			args: args{
				orderName: "p_test2_1",
			},
			want: "p_test2_1",
		},
		{
			name: "project",
			args: args{
				orderName: "p_015a3fbd6ae04f9ab6132d9cee5b99d5_0",
			},
			want: "p_015a3f",
		},
		{
			name: "application",
			args: args{
				orderName: "a_015a3fbd6ae04f9ab6132d9cee5b99d5_1",
			},
			want: "a_015a3f_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDeploymentOrderShowName(tt.args.orderName)
			assert.Equal(t, got, tt.want)
		})
	}
}
