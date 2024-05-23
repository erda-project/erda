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

package client

import (
	"context"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"

	"github.com/erda-project/erda/internal/tools/cluster-agent/config"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func Test_getClusterInfo(t *testing.T) {
	fakeSa := "cluster-agent"

	defer monkey.UnpatchAll()
	monkey.Patch(k8sclient.NewForInCluster,
		func(...k8sclient.Option) (*k8sclient.K8sClient, error) {
			return &k8sclient.K8sClient{
				ClientSet: fakeclientset.NewSimpleClientset(&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: fakeSa,
					},
					Secrets: []corev1.ObjectReference{
						{
							Name: "cluster-agent-token-mvp6d",
						},
					},
				}, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-agent-token-mvp6d",
					},
					Data: map[string][]byte{
						caCrtKey:    []byte("fake ca data"),
						tokenSecKey: []byte("fake token data"),
					},
				}),
			}, nil
		},
	)
	c := New(WithConfig(&config.Config{
		CollectClusterInfo: true,
		ServiceAccount:     fakeSa,
	}))
	_, err := c.loadClusterInfo(context.Background())
	assert.NoError(t, err)
}
