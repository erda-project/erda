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

package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_marshalQualityResult(t *testing.T) {
	type args struct {
		qr apistructs.QualityGateResult
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				qr: apistructs.QualityGateResult{
					Status: "ok",
					Conditions: []apistructs.QualityGateConditionResult{
						{
							MetricKey: "coverage",
							Status:    "ok",
						},
					},
				},
			},
			want: `{"status":"ok","conditions":[{"status":"ok","metricKey":"coverage","comparator":"","errorThreshold":"","actualValue":""}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalQualityResult(tt.args.qr)
			assert.NoError(t, err)
			if got != tt.want {
				t.Errorf("marshalQualityResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
