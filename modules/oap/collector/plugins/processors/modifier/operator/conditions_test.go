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
	"testing"
)

func TestKeyExist_Match(t *testing.T) {
	type fields struct {
		cfg ConditionCfg
	}
	type args struct {
		pairs map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			fields: fields{cfg: ConditionCfg{
				Key: "aaa",
				Op:  "key_exist",
			}},
			args: args{pairs: map[string]interface{}{"aaa": ""}},
			want: true,
		},
		{
			fields: fields{cfg: ConditionCfg{
				Key: "aaa",
				Op:  "key_exist",
			}},
			args: args{pairs: map[string]interface{}{"bbb": ""}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KeyExist{
				cfg: tt.fields.cfg,
			}
			if got := k.Match(tt.args.pairs); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueMatch_Match(t *testing.T) {
	type fields struct {
		cfg ConditionCfg
	}
	type args struct {
		pairs map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			fields: fields{
				cfg: ConditionCfg{
					Key:   "pod_namespace",
					Value: "addon(.*?)",
				},
			},
			args: args{pairs: map[string]interface{}{
				"pod_namespace": "group-addon-rocketmq--w28b7279764b64180843611a3dfbe0f4b",
			}},
			want: true,
		},
		{
			fields: fields{
				cfg: ConditionCfg{
					Key:   "pod_namespace",
					Value: "addon(.*?)",
				},
			},
			args: args{pairs: map[string]interface{}{
				"pod_namespace": "addon-redis--yaf54cd190f484ddba84c681e4a9eba69",
			}},
			want: true,
		},
		{
			fields: fields{
				cfg: ConditionCfg{
					Key:   "pod_namespace",
					Value: "addon(.*?)",
				},
			},
			args: args{pairs: map[string]interface{}{
				"pod_namespace": "app_name",
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueMatch(tt.fields.cfg)
			if got := v.Match(tt.args.pairs); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
