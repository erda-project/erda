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

package formatter

import (
	"context"
	"net/http"
	"testing"

	"github.com/rancher/apiserver/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/erda-project/erda/modules/cmp/cache"
)

type podInterface struct {
	v1.PodInterface
}

func (p *podInterface) List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error) {
	cpu100m, err := resource.ParseQuantity("100m")
	if err != nil {
		return nil, err
	}
	cpu1, err := resource.ParseQuantity("1")
	if err != nil {
		return nil, err
	}

	mem500Mi, err := resource.ParseQuantity("500Mi")
	if err != nil {
		return nil, err
	}
	mem1Gi, err := resource.ParseQuantity("1Gi")
	if err != nil {
		return nil, err
	}

	return &corev1.PodList{
		Items: []corev1.Pod{
			{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpu100m,
									corev1.ResourceMemory: mem500Mi,
								},
							},
						},
					},
				},
			},
			{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpu1,
									corev1.ResourceMemory: mem1Gi,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func TestGetNodeAllocatedRes(t *testing.T) {
	cache, err := cache.New(32<<10, 16<<10)
	if err != nil {
		t.Error(err)
	}

	nodeFormatter := &NodeFormatter{
		ctx:       context.WithValue(context.Background(), "clusterName", "test"),
		podClient: &podInterface{},
		podsCache: cache,
	}

	res, err := nodeFormatter.getNodeAllocatedRes(nodeFormatter.ctx, "")
	if err != nil {
		t.Error(err)
	}
	cpu := res["CPU"].(int64)
	mem := res["Memory"].(int64)
	pods := res["Pods"].(int64)
	if cpu != 1100 {
		t.Errorf("test failed, expected cpu 1100, actual %d", cpu)
	}
	if mem != 1598029824 {
		t.Errorf("test failed, expected 1598029824, actual %d", mem)
	}
	if pods != 2 {
		t.Errorf("test failed, expected Pods 200, actual %d", pods)
	}
}

func TestNodeFormatter_Formatter(t *testing.T) {
	res := &types.RawResource{
		APIObject: types.APIObject{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"allocatable": map[string]interface{}{
						"cpu":    "8",
						"memory": "16Gi",
						"pods":   "200",
					},
					"capacity": map[string]interface{}{
						"cpu":    "8",
						"memory": "16Gi",
						"pods":   "200",
					},
				},
			},
		},
	}

	podCache, err := cache.New(1<<20, 1<<10)
	if err != nil {
		t.Errorf("test failed, failed to create cache, %v", err)
	}

	n := &NodeFormatter{
		ctx:       context.WithValue(context.Background(), "clusterName", "test"),
		podClient: &podInterface{},
		podsCache: podCache,
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("test failed, failed to create a http request, %v", err)
	}
	request := &types.APIRequest{
		Request: req,
	}
	n.Formatter(request, res)
	n.Formatter(request, res)
}
