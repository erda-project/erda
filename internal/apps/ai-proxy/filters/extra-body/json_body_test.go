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

package extra_body

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

func compare_map(got, want map[string]any) bool {
	gotStr, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantStr, err := json.Marshal(want)
	if err != nil {
		return false
	}
	return string(gotStr) == string(wantStr)
}

func Test_unmarshalJSONFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "empty {}",
			args: args{
				s: `{}`,
			},
			want:    map[string]any{},
			wantErr: false,
		},
		{
			name: "extra json body for non-stream",
			args: args{
				s: `{"enable_thinking": true, "thinking_budget": 31334}`,
			},
			want:    map[string]any{"enable_thinking": true, "thinking_budget": 31334},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalJSONFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalJSONFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compare_map(got, tt.want) {
				t.Errorf("unmarshalJSONFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExtraJSONBody(t *testing.T) {
	type args struct {
		m        *metadata.Metadata
		isStream bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "extra json body for non-stream",
			args: args{
				m: &metadata.Metadata{
					Public: map[string]string{
						"extra_json_body_for_non_stream": `{"enable_thinking": false}`,
					},
				},
				isStream: false,
			},
			want:    map[string]any{"enable_thinking": false},
			wantErr: false,
		},
		{
			name: "extra json body for stream",
			args: args{
				m: &metadata.Metadata{
					Public: map[string]string{
						"extra_json_body_for_stream": `{"enable_thinking": true, "thinking_budget": 31334}`,
					},
				},
				isStream: true,
			},
			want:    map[string]any{"thinking_budget": 31334, "enable_thinking": true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetExtraJSONBody(tt.args.m, tt.args.isStream)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExtraJSONBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compare_map(got, tt.want) {
				t.Errorf("GetExtraJSONBody() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFulfillExtraJSONBody(t *testing.T) {
	type args struct {
		m        *metadata.Metadata
		isStream bool
		body     map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "add new extra json body",
			args: args{
				m: &metadata.Metadata{
					Public: map[string]string{
						"extra_json_body_for_non_stream": `{"enable_thinking": false}`,
					},
				},
				isStream: false,
				body:     map[string]any{},
			},
			want:    map[string]any{"enable_thinking": false},
			wantErr: false,
		},
		{
			name: "override extra json body",
			args: args{
				m: &metadata.Metadata{
					Public: map[string]string{
						"extra_json_body_for_non_stream": `{"enable_thinking": false}`,
					},
				},
				isStream: false,
				body:     map[string]any{"enable_thinking": true},
			},
			want:    map[string]any{"enable_thinking": false},
			wantErr: false,
		},
		{
			name: "stream type mismatch",
			args: args{
				m: &metadata.Metadata{
					Public: map[string]string{
						"extra_json_body_for_non_stream": `{"enable_thinking": false}`,
					},
				},
				isStream: true,
				body:     map[string]any{"thinking_budget": 31334},
			},
			want:    map[string]any{"thinking_budget": 31334},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FulfillExtraJSONBody(tt.args.m, tt.args.isStream, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("FulfillExtraJSONBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
