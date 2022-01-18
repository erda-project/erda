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

package pod

import (
	"reflect"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/modules/oap/collector/plugins/processors/k8s-tagger/metadata"
)

func TestCache_extractPodMetadata(t *testing.T) {
	type fields struct {
		aInclude, lInclude []string
	}
	type args struct {
		pod *apiv1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			args: args{pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "aaa",
					Namespace:   "default",
					UID:         "aaa-bbb-ccc-ddd",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: apiv1.PodSpec{},
				Status: apiv1.PodStatus{
					PodIP: "1.1.1.1",
				},
			}},
			want: map[string]string{
				metadata.PrefixPod + "name":      "aaa",
				metadata.PrefixPod + "namespace": "default",
				metadata.PrefixPod + "uid":       "aaa-bbb-ccc-ddd",
				metadata.PrefixPod + "ip":        "1.1.1.1",
			},
		},
		{
			name: "annotations&labels",
			fields: fields{
				aInclude: []string{"msp.erda.cloud/(.+)"},
				lInclude: []string{"(.+).erda.cloud/(.+)"},
			},
			args: args{pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aaa",
					Namespace: "default",
					UID:       "aaa-bbb-ccc-ddd",
					Labels: map[string]string{
						"dop.erda.cloud/a": "b",
						"msp.erda.cloud/c": "d",
					},
					Annotations: map[string]string{
						"dop.erda.cloud/a": "b",
						"msp.erda.cloud/e": "f",
					},
				},
				Spec: apiv1.PodSpec{},
				Status: apiv1.PodStatus{
					PodIP: "1.1.1.1",
				},
			}},
			want: map[string]string{
				metadata.PrefixPod + "name":                        "aaa",
				metadata.PrefixPod + "namespace":                   "default",
				metadata.PrefixPod + "uid":                         "aaa-bbb-ccc-ddd",
				metadata.PrefixPod + "ip":                          "1.1.1.1",
				metadata.PrefixPodLabels + "dop_erda_cloud_a":      "b",
				metadata.PrefixPodLabels + "msp_erda_cloud_c":      "d",
				metadata.PrefixPodAnnotations + "msp_erda_cloud_e": "f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCache(nil, tt.fields.aInclude, tt.fields.lInclude)
			if got := c.extractPodMetadata(tt.args.pod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPodMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
