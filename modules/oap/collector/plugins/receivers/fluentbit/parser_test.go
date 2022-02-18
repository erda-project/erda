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

package fluentbit

import (
	"reflect"
	"testing"

	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
	"github.com/stretchr/testify/assert"
)

func Test_parseItem(t *testing.T) {
	type args struct {
		value []byte
		cfg   flbKeyMappings
	}
	tests := []struct {
		name    string
		args    args
		want    *lpb.Log
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				value: []byte("{\"date\":\"2018-05-30T09:39:52.000681Z\",\"log\":\"hello world\\n\",\"stream\":\"stdout\",\"time\":\"2021-08-16T07:56:01.025279973Z\",\"level\":\"INFO\",\"source\":\"container\",\"kubernetes\":{\"pod_name\":\"pod\",\"namespace\":\"ns\",\"container_name\":\"cname\",\"container_id\":\"cid\",\"pod_id\":\"pid\",\"labels\":{\"k\":\"v\"},\"annotations\":{\"k\":\"v\"}}}"),
				cfg: flbKeyMappings{
					TimeUnixNano: "time",
					Name:         "source",
					Content:      "log",
					Severity:     "level",
					Kubernetes:   "kubernetes",
				},
			},
			want: &lpb.Log{
				TimeUnixNano: 1629100561025279973,
				Name:         "container",
				Severity:     "INFO",
				Attributes: map[string]string{
					"k8s_annotations_k":  "v",
					"k8s_container_id":   "cid",
					"k8s_container_name": "cname",
					"k8s_labels_k":       "v",
					"k8s_namespace":      "ns",
					"k8s_pod_id":         "pid",
					"k8s_pod_name":       "pod",
				},
				Content: "hello world\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseItem(tt.args.value, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseMapStr(t *testing.T) {
	type args struct {
		prefix string
		data   []byte
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			args: args{
				prefix: "k8s",
				data:   []byte("{\"pod_name\":\"pod\",\"namespace\":\"ns\",\"container_name\":\"cname\",\"container_id\":\"cid\",\"pod_id\":\"pid\",\"labels\":{\"k\":\"v\"},\"annotations\":{\"k\":\"v\"}}"),
			},
			want: map[string]string{
				"k8s_annotations_k":  "v",
				"k8s_container_id":   "cid",
				"k8s_container_name": "cname",
				"k8s_labels_k":       "v",
				"k8s_namespace":      "ns",
				"k8s_pod_id":         "pid",
				"k8s_pod_name":       "pod",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMapStr(tt.args.prefix, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMapStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
