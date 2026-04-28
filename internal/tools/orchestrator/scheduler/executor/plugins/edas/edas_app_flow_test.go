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
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
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
				"core.erda.cloud/service-name":    "test-service",
				"core.erda.cloud/servicegroup-id": "test-sg-id",
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
				"core.erda.cloud/service-name":    "test-service",
				"core.erda.cloud/servicegroup-id": "",
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
				"core.erda.cloud/service-name":    "",
				"core.erda.cloud/servicegroup-id": "test-sg-id",
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
				"core.erda.cloud/service-name":    "test-service-with-dashes",
				"core.erda.cloud/servicegroup-id": "test-sg-id-with-dashes_and_underscores",
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
	assert.Equal(t, "test-service", labels["core.erda.cloud/service-name"])
	assert.Equal(t, "test-sg-id", labels["core.erda.cloud/servicegroup-id"])

	// Verify JSON structure
	expectedJSON := `{"core.erda.cloud/service-name":"test-service","core.erda.cloud/servicegroup-id":"test-sg-id"}`
	var expectedLabels map[string]string
	err = json.Unmarshal([]byte(expectedJSON), &expectedLabels)
	require.NoError(t, err)
	assert.Equal(t, expectedLabels, labels)
}

func TestTrySendErrSendsWhenReceiverIsAvailable(t *testing.T) {
	errChan := make(chan error, 1)
	expectedErr := assert.AnError

	trySendErr(errChan, expectedErr)
	require.ErrorIs(t, <-errChan, expectedErr)
}

func TestTrySendErrDoesNotBlockWhenChannelIsFull(t *testing.T) {
	errChan := make(chan error, 1)
	errChan <- assert.AnError

	trySendErr(errChan, assert.AnError)
	require.ErrorIs(t, <-errChan, assert.AnError)
}

func TestRunServiceBatchCancelsOnFirstError(t *testing.T) {
	cancelableStarted := make(chan struct{})
	cancelObserved := make(chan struct{})
	batch := []*apistructs.Service{
		{Name: "failed"},
		{Name: "cancelable"},
	}

	err := runServiceBatch(context.Background(), batch, func(ctx context.Context, svc *apistructs.Service) error {
		switch svc.Name {
		case "failed":
			<-cancelableStarted
			return assert.AnError
		case "cancelable":
			close(cancelableStarted)
			<-ctx.Done()
			close(cancelObserved)
			return nil
		default:
			return nil
		}
	})

	require.ErrorIs(t, err, assert.AnError)
	select {
	case <-cancelObserved:
	case <-time.After(time.Second):
		t.Fatal("cancelable service did not observe cancellation")
	}
}

func TestRunServiceBatchReturnsParentContextErrorWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := runServiceBatch(ctx, []*apistructs.Service{{Name: "svc"}}, func(ctx context.Context, svc *apistructs.Service) error {
		called = true
		return nil
	})

	require.ErrorIs(t, err, context.Canceled)
	require.False(t, called, "run should not be called when parent context is already canceled")
}

func TestResolveServiceSelectorFromDeployment(t *testing.T) {
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"match": "deploy",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						types.LabelServiceName:    "deploy-service",
						types.LabelServiceGroupID: "deploy-group",
						types.LabelEDASAppID:      "deploy-app-id",
					},
				},
			},
		},
	}

	assert.Equal(t, map[string]string{
		types.LabelServiceName:    "deploy-service",
		types.LabelServiceGroupID: "deploy-group",
	},
		resolveServiceSelectorFromDeployment(deployment, "test-sg", "test-service"))

	assert.Equal(t, map[string]string{
		types.LabelServiceName:    "test-service",
		types.LabelServiceGroupID: "test-sg",
	}, resolveServiceSelectorFromDeployment(nil, "test-sg", "test-service"))

	deployment.Spec.Template.Labels = map[string]string{
		types.LabelEDASAppID: "deploy-app-id",
	}
	assert.Equal(t, map[string]string{
		types.LabelEDASAppID: "deploy-app-id",
	}, resolveServiceSelectorFromDeployment(deployment, "test-sg", "test-service"))

	deployment.Spec.Template.Labels = map[string]string{
		"custom": "deploy-value",
	}
	deployment.Spec.Selector.MatchLabels = map[string]string{
		"match": "deploy",
	}
	assert.Equal(t, map[string]string{
		"match": "deploy",
	}, resolveServiceSelectorFromDeployment(deployment, "test-sg", "test-service"))
}
