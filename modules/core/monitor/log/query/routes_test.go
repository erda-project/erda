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

package query

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFilename(t *testing.T) {
	type args struct {
		r    *RequestCtx
		meta *LogMeta
	}
	tests := []struct {
		name  string
		args  args
		match string
	}{
		{
			name: "normal",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "",
					Stream:        "stdout",
					Start:         0,
					End:           0,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
				meta: &LogMeta{
					Source: "container",
					ID:     "",
					Tags: map[string]string{
						"pod_name":              "aaa-default-xxx",
						"dice_application_name": "app1",
						"dice_service_name":     "svc1",
					},
				},
			},
			match: "svc1_stdout_\\d+\\.log",
		},
		{
			name: "no service name",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "",
					Stream:        "stdout",
					Start:         0,
					End:           0,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
				meta: &LogMeta{
					Source: "container",
					ID:     "",
					Tags: map[string]string{
						"pod_name":              "aaa-default-xxx",
						"dice_application_name": "app1",
					},
				},
			},
			match: "app1_stdout_\\d+\\.log",
		},
		{
			name: "no app name",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "",
					Stream:        "stdout",
					Start:         0,
					End:           0,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
				meta: &LogMeta{
					Source: "container",
					ID:     "",
					Tags: map[string]string{
						"pod_name": "aaa-default-xxx",
					},
				},
			},
			match: "aaa-default-xxx_stdout_\\d+\\.log",
		},
		{
			name: "no pod name",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "aaa",
					Stream:        "stdout",
					Start:         0,
					End:           0,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
				meta: &LogMeta{
					Source: "container",
					ID:     "aaa",
					Tags:   map[string]string{},
				},
			},
			match: "aaa_stdout_\\d+\\.log",
		},
		{
			name: "no meta",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "aaa",
					Stream:        "stdout",
					Start:         0,
					End:           0,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
				meta: nil,
			},
			match: "aaa_stdout_\\d+\\.log",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := regexp.Match(tt.match, []byte(getFilename(tt.args.r, tt.args.meta)))
			assert.Nil(t, err)
			assert.True(t, ok)
		})
	}
}

func Test_normalizeRequest(t *testing.T) {
	type args struct {
		r *RequestCtx
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		exceptedCtx *RequestCtx
	}{
		{
			name: "normal",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "aaa",
					Stream:        "stdout",
					Start:         1627266816000000000,
					End:           1627266900000000000,
					Count:         -200,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
			},
			wantErr: false,
			exceptedCtx: &RequestCtx{
				RequestID:     "",
				LogID:         "",
				Source:        "container",
				ID:            "aaa",
				Stream:        "stdout",
				Start:         1627266816000000000,
				End:           1627266900000000000,
				Count:         -200,
				ApplicationID: "app1",
				ClusterName:   "org1",
			},
		},
		{
			name: "default settings",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "aaa",
					Stream:        "",
					Start:         0,
					End:           1627266900000000000,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
			},
			wantErr: false,
			exceptedCtx: &RequestCtx{
				RequestID:     "",
				LogID:         "",
				Source:        "container",
				ID:            "aaa",
				Stream:        "stdout",
				Start:         1626662100000000000,
				End:           1627266900000000000,
				Count:         50,
				ApplicationID: "app1",
				ClusterName:   "org1",
			},
		},
		{
			name: "invalid ctx: no source",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "",
					ID:            "aaa",
					Stream:        "",
					Start:         0,
					End:           1627266900000000000,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
			},
			wantErr: true,
			exceptedCtx: &RequestCtx{
				RequestID:     "",
				LogID:         "",
				Source:        "",
				ID:            "aaa",
				Stream:        "",
				Start:         0,
				End:           1627266900000000000,
				Count:         0,
				ApplicationID: "app1",
				ClusterName:   "org1",
			},
		},
		{
			name: "invalid ctx: no id",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "",
					Stream:        "",
					Start:         0,
					End:           1627266900000000000,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
			},
			wantErr: true,
			exceptedCtx: &RequestCtx{
				RequestID:     "",
				LogID:         "",
				Source:        "container",
				ID:            "",
				Stream:        "",
				Start:         0,
				End:           1627266900000000000,
				Count:         0,
				ApplicationID: "app1",
				ClusterName:   "org1",
			},
		},
		{
			name: "invalid ctx: impossible start&end",
			args: args{
				r: &RequestCtx{
					RequestID:     "",
					LogID:         "",
					Source:        "container",
					ID:            "aaa",
					Stream:        "stdout",
					Start:         1627267900000000000,
					End:           1627266900000000000,
					Count:         0,
					ApplicationID: "app1",
					ClusterName:   "org1",
				},
			},
			wantErr: true,
			exceptedCtx: &RequestCtx{
				RequestID:     "",
				LogID:         "",
				Source:        "container",
				ID:            "aaa",
				Stream:        "stdout",
				Start:         1627267900000000000,
				End:           1627266900000000000,
				Count:         0,
				ApplicationID: "app1",
				ClusterName:   "org1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := normalizeRequest(tt.args.r)
			if !tt.wantErr {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.exceptedCtx, tt.args.r)
		})
	}
}

func Test_provider_checkLogMeta(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		source string
		id     string
		key    string
		value  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "non container log",
			args: args{
				source: "pipeline-job-123",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				source: "container",
				id:     "aaa",
				key:    "dice_application_id",
				value:  "app-1",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "trigger error",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						errorTrigger: true,
					},
				},
			},
			args: args{
				source: "container",
				id:     "aaa",
				key:    "dice_application_id",
				value:  "app-1",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "normal but empty",
			fields: fields{
				p: &provider{
					cqlQuery: &mockCqlQuery{
						emptyResult: true,
					},
				},
			},
			args: args{
				source: "container",
				id:     "aaa",
				key:    "dice_application_id",
				value:  "app-1",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := mockProvider()
			if tt.fields.p != nil {
				mp = tt.fields.p
			}
			got, err := mp.checkLogMeta(tt.args.source, tt.args.id, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkLogMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkLogMeta() got = %v, want %v", got, tt.want)
			}
		})
	}
}
