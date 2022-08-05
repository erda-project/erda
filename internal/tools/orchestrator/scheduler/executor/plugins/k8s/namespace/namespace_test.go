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

package namespace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func TestNamespace(t *testing.T) {
	ns := New(WithKubernetesClient(fakeclientset.NewSimpleClientset()))

	assert.NoError(t, ns.Create("fake-namespace", map[string]string{
		"hello": "world",
	}))
	assert.Equal(t, ns.Exists("fake-namespace"), nil)

	type args struct {
		namespace string
		labels    map[string]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update labels",
			args: args{
				namespace: "fake-namespace",
				labels: map[string]string{
					"hello": "new-world",
				},
			},
		},
		{
			name: "non-existent namespace",
			args: args{
				namespace: "non-existent-namespace",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ns.Update(tt.args.namespace, tt.args.labels); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	newNs, err := ns.cs.CoreV1().Namespaces().Get(context.Background(), "fake-namespace", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"hello": "new-world"}, newNs.Labels)
	assert.NoError(t, ns.Delete("fake-namespace", true))
}
