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

package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func Test_caclFineGrainedCPU(t *testing.T) {
	k := &Kubernetes{}

	testCases := []struct {
		name       string
		requestCPU float64
		maxCPU     float64
		ratio      float64
		wantCPU    float64
		wantMaxCPU float64
		wantErr    bool
	}{
		{
			name:       "valid input with maxCPU set",
			requestCPU: 1,
			maxCPU:     2,
			ratio:      10,
			wantCPU:    0.1,
			wantMaxCPU: 2,
			wantErr:    false,
		},
		{
			name:       "valid input without maxCPU",
			requestCPU: 1,
			maxCPU:     0,
			ratio:      5,
			wantCPU:    0.2,
			wantMaxCPU: 1,
			wantErr:    false,
		},
		{
			name:       "requestCPU less than MIN_CPU_SIZE",
			requestCPU: 0.05,
			maxCPU:     2.0,
			ratio:      1.5,
			wantCPU:    0,
			wantMaxCPU: 0,
			wantErr:    true,
		},
		{
			name:       "maxCPU less than requestCPU",
			requestCPU: 2.0,
			maxCPU:     1.0,
			ratio:      1.5,
			wantCPU:    0,
			wantMaxCPU: 0,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualCPU, actualMaxCPU, err := k.calcFineGrainedCPU(tc.requestCPU, tc.maxCPU, tc.ratio)

			// Assert CPU values
			assert.Equal(t, tc.wantCPU, actualCPU)
			assert.Equal(t, tc.wantMaxCPU, actualMaxCPU)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetWorkspaceRatio(t *testing.T) {
	tests := []struct {
		name          string
		options       map[string]string
		workspace     string
		expectedValue float64
	}{
		{
			name:          "Default ratio when no options provided",
			options:       map[string]string{},
			expectedValue: DefaultRatio,
		},
		{
			name:          "Set ratio for production",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10"},
			workspace:     apistructs.ProdWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Non-production workspace",
			options:       map[string]string{"DEV_CPU_SUBSCRIBE_RATIO": "10"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Non-production workspace overrides global",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10", "DEV_CPU_SUBSCRIBE_RATIO": "20"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 20,
		},
		{
			name:          "Ratio < 1.0",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10", "DEV_CPU_SUBSCRIBE_RATIO": "0.5"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Ratio < 1.0 case2",
			options:       map[string]string{"DEV_CPU_SUBSCRIBE_RATIO": "0.5"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value float64
			getWorkspaceRatio(tt.options, tt.workspace, "CPU", &value)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
