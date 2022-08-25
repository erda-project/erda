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
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	fakeclientset "k8s.io/client-go/kubernetes/fake"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func Test_IPToHostname(t *testing.T) {
	type fields struct {
		K8sClient func() kubernetes.Interface
	}

	type args struct {
		IP string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "case 1",
			fields: fields{
				K8sClient: func() kubernetes.Interface {
					return fakeclientset.NewSimpleClientset(&corev1.NodeList{
						Items: []corev1.Node{
							genFakeNode(t, "192.168.0.1"),
							genFakeNode(t, "192.168.0.2"),
							genFakeNode(t, "192.168.0.3"),
						},
					})
				},
			},
			args: args{
				IP: "192.168.0.2",
			},
			want: "node-192168000002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Kubernetes{
				k8sClient: &k8sclient.K8sClient{
					ClientSet: tt.fields.K8sClient(),
				},
			}
			if got := k.IPToHostname(tt.args.IP); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IPToHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SetNodeLabel(t *testing.T) {
	k := Kubernetes{
		k8sClient: &k8sclient.K8sClient{
			ClientSet: fakeclientset.NewSimpleClientset(&corev1.NodeList{
				Items: []corev1.Node{
					genFakeNode(t, "192.168.0.1"),
					genFakeNode(t, "192.168.0.2"),
					genFakeNode(t, "192.168.0.3"),
					genFakeNode(t, "192.168.0.4"),
					genFakeNode(t, "192.168.0.5"),
				},
			}),
		},
	}

	type args struct {
		hosts  []string
		labels map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				hosts: []string{
					"192.168.0.2",
					"192.168.0.3",
				},
				labels: map[string]string{
					"dice/org-erda":       "true",
					"dice/platform":       "true",
					"dice/workspace-test": "true",
					"other":               "new",
				},
			},
			want: map[string]string{
				"dice/platform":         "true",
				"dice/workspace-test":   "true",
				"dice/other":            "new",
				"beta.kubernetes.io/os": "linux",
				"dice/org-erda":         "true",
				"other":                 "old",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.SetNodeLabels(executortypes.NodeLabelSetting{}, tt.args.hosts, tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetNodeLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, host := range tt.args.hosts {
				node, err := k.k8sClient.ClientSet.CoreV1().Nodes().Get(context.Background(), convertNodeName(t, host),
					metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(node.Labels, tt.want) {
					t.Errorf("SetNodeLabels() got = %v, want %v", node.Labels, tt.want)
				}
			}
		})
	}

}

func genFakeNode(t *testing.T, ip string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: convertNodeName(t, ip),
			Labels: map[string]string{
				"beta.kubernetes.io/os": "linux",
				"dice/org-erda":         "true",
				"other":                 "old",
			},
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeInternalIP,
					Address: ip,
				},
				{
					Type:    corev1.NodeExternalIP,
					Address: "fake",
				},
			},
		},
	}
}

func convertNodeName(t *testing.T, ip string) string {
	var formatIP string
	for _, r := range strings.Split(ip, ".") {
		i, err := strconv.Atoi(r)
		if err != nil {
			t.Fatal(err)
		}
		formatIP += fmt.Sprintf("%03d", i)
	}

	return fmt.Sprintf("node-%s", formatIP)
}
