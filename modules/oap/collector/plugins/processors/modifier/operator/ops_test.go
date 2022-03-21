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

package operator

import (
	"reflect"
	"testing"
)

func TestTrimPrefix_Operate(t1 *testing.T) {
	type fields struct {
		cfg ModifierCfg
	}
	type args struct {
		pairs map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{
			name: "",
			fields: fields{
				cfg: ModifierCfg{
					Key:    "kubernetes_",
					Value:  "",
					Action: "trim_prefix",
				},
			},
			args: args{pairs: map[string]interface{}{
				"kubernetes_pod_ip": "1.1.1.1",
			}},
			want: map[string]interface{}{
				"pod_ip": "1.1.1.1",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TrimPrefix{
				cfg: tt.fields.cfg,
			}
			if got := t.Operate(tt.args.pairs); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Operate() = %v, want %v", got, tt.want)
			}
		})
	}
}
