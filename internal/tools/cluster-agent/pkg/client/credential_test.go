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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_getRetryWatcher(t *testing.T) {
	c := New()
	fc := &fakeclientset.Clientset{}
	fc.AddReactor("list", "secrets", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &corev1.SecretList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "1",
			},
			Items: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      apistructs.ErdaClusterCredential,
						Namespace: metav1.NamespaceDefault,
					},
					Data: map[string][]byte{
						apistructs.ClusterAccessKey: []byte("test"),
					},
				},
			},
		}, nil
	})
	_, err := c.getRetryWatcher(fc, metav1.NamespaceDefault)
	assert.NoError(t, err)
}
