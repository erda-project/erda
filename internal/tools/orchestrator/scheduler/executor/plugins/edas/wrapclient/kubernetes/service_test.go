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

package kubernetes

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
)

func TestCreateOrUpdateK8sService(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	kubernetesWrapper := wrapKubernetes{
		l:         logrus.WithField("unit-test", "wrap-kubernetes"),
		namespace: metav1.NamespaceDefault,
		cs:        fakeClient,
	}

	appName := "test-app"
	appID := "test-app-id"
	ports := []int{80, 8080}

	args := struct {
		appName string
		appID   string
		ports   []int
	}{
		appName: appName,
		appID:   appID,
		ports:   ports,
	}

	tests := []struct {
		name          string
		existingSvc   bool
		expectedError bool
	}{
		{
			name:          "Create new service",
			existingSvc:   false,
			expectedError: false,
		},
		{
			name:          "Update existing service",
			existingSvc:   true,
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.existingSvc {
				_ = kubernetesWrapper.CreateK8sService(appName, appID, ports)
			}

			err := kubernetesWrapper.CreateOrUpdateK8sService(args.appName, args.appID, args.ports)

			if test.expectedError && err == nil {
				t.Error("Expected an error, but got nil.")
			} else if !test.expectedError && err != nil {
				t.Errorf("Expected no error, but got an error: %v", err)
			}
		})
	}
}

func TestCombineK8sService(t *testing.T) {
	wrapCs := wrapKubernetes{}
	// Test case 1
	appName := "my-app"
	appID := "123"
	ports := []int{8080, 9090}

	expectedService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appName,
			Labels: make(map[string]string),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				types.EDASAppIDLabel: appID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http-0",
					Port:       int32(8080),
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "http-1",
					Port:       int32(9090),
					TargetPort: intstr.FromInt(9090),
				},
			},
		},
	}

	service := wrapCs.combineK8sService(appName, appID, ports)

	// Compare the expected service with the actual service
	if !reflect.DeepEqual(service, expectedService) {
		t.Errorf("CombineK8sService() = %v, want %v", service, expectedService)
	}

}
