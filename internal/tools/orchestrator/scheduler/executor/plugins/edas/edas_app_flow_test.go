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

package edas

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
)

func TestSetLabels(t *testing.T) {
	tests := []struct {
		name        string
		svcSpec     *types.ServiceSpec
		sgID        string
		serviceName string
		expected    map[string]string
		expectError bool
	}{
		{
			name: "successful label setting",
			svcSpec: &types.ServiceSpec{
				Name: "test-service",
			},
			sgID:        "test-sg-id",
			serviceName: "test-service",
			expected: map[string]string{
				"app":             "test-service",
				"servicegroup-id": "test-sg-id",
			},
			expectError: false,
		},
		{
			name:        "nil service spec",
			svcSpec:     nil,
			sgID:        "test-sg-id",
			serviceName: "test-service",
			expectError: true,
		},
		{
			name: "empty service group ID",
			svcSpec: &types.ServiceSpec{
				Name: "test-service",
			},
			sgID:        "",
			serviceName: "test-service",
			expected: map[string]string{
				"app":             "test-service",
				"servicegroup-id": "",
			},
			expectError: false,
		},
		{
			name: "empty service name",
			svcSpec: &types.ServiceSpec{
				Name: "test-service",
			},
			sgID:        "test-sg-id",
			serviceName: "",
			expected: map[string]string{
				"app":             "",
				"servicegroup-id": "test-sg-id",
			},
			expectError: false,
		},
		{
			name: "special characters in labels",
			svcSpec: &types.ServiceSpec{
				Name: "test-service",
			},
			sgID:        "test-sg-id-with-dashes_and_underscores",
			serviceName: "test-service-with-dashes",
			expected: map[string]string{
				"app":             "test-service-with-dashes",
				"servicegroup-id": "test-sg-id-with-dashes_and_underscores",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setLabels(tt.svcSpec, tt.sgID, tt.serviceName)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.svcSpec)

			// Parse the labels JSON
			var actualLabels map[string]string
			err = json.Unmarshal([]byte(tt.svcSpec.Labels), &actualLabels)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, actualLabels)
		})
	}
}

func TestSetLabels_JSONMarshaling(t *testing.T) {
	svcSpec := &types.ServiceSpec{
		Name: "test-service",
	}

	err := setLabels(svcSpec, "test-sg-id", "test-service")
	require.NoError(t, err)

	// Verify that the JSON is valid and properly formatted
	var labels map[string]string
	err = json.Unmarshal([]byte(svcSpec.Labels), &labels)
	require.NoError(t, err)

	// Verify specific labels are set
	assert.Equal(t, "test-service", labels["app"])
	assert.Equal(t, "test-sg-id", labels["servicegroup-id"])

	// Verify JSON structure
	expectedJSON := `{"app":"test-service","servicegroup-id":"test-sg-id"}`
	var expectedLabels map[string]string
	err = json.Unmarshal([]byte(expectedJSON), &expectedLabels)
	require.NoError(t, err)
	assert.Equal(t, expectedLabels, labels)
}
