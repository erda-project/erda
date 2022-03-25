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
	"regexp"
	"testing"
)

func TestRegex_Operate(t *testing.T) {
	type fields struct {
		cfg     ModifierCfg
		pattern *regexp.Regexp
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
			fields: fields{
				cfg: ModifierCfg{
					Key:   "id",
					Value: "^\\/kubepods\\/\\w+\\/[\\w|\\-]+\\/(?P<container_id>\\w+)",
				},
			},
			args: args{pairs: map[string]interface{}{"id": "/kubepods/burstable/poda102e534-399d-4553-a523-e6489222ca96/4c949b590a29dd49b08c2576cb408c57c69a58142aae9d75a7c79ca08dbaf7b9"}},
			want: map[string]interface{}{
				"id":           "/kubepods/burstable/poda102e534-399d-4553-a523-e6489222ca96/4c949b590a29dd49b08c2576cb408c57c69a58142aae9d75a7c79ca08dbaf7b9",
				"container_id": "4c949b590a29dd49b08c2576cb408c57c69a58142aae9d75a7c79ca08dbaf7b9",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegex(tt.fields.cfg)
			if got := r.Operate(tt.args.pairs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Operate() = %v, want %v", got, tt.want)
			}
		})
	}
}
