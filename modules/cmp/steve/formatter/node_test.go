// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package formatter

import (
	"context"
	"testing"

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
		ctx:       context.Background(),
		podClient: &podInterface{},
		podsCache: cache,
	}

	res, err := nodeFormatter.getNodeAllocatedRes(nodeFormatter.ctx, "")
	if err != nil {
		t.Error(err)
	}
	if res.CPU != 1100 {
		t.Errorf("test failed, expected cpm 1100, actual %d", res.CPU)
	}
	if res.Memory != 1598029824 {
		t.Errorf("test failed, expected 1598029824, actual %d", res.Memory)
	}
	if res.Pods != 2 {
		t.Errorf("test failed, expected Pods 200, actual %d", res.Pods)
	}
}
