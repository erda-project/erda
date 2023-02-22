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

package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/core/monitor/settings/pb"
)

func Test_insertOrgFilter(t *testing.T) {
	type args struct {
		typ     string
		orgID   string
		orgName string
		filters string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test_insertOrgFilter",
			args: args{
				typ:     "metric",
				orgID:   "1",
				orgName: "terminus",
				filters: `[{"key":"erda","value":"pjy"}]`,
			},
			want: `[{"key":"org_name","value":"terminus"},{"key":"erda","value":"pjy"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := insertOrgFilter(tt.args.typ, tt.args.orgID, tt.args.orgName, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("insertOrgFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("insertOrgFilter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getConfigFromDays(t *testing.T) {
	type args struct {
		days ttl
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_getConfigFromDays",
			args: args{
				ttl{
					TTL: 3,
				},
			},
			want: `{"ttl":"72h0m0s","hot_ttl":"0s"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConfigFromDays(tt.args.days); got != tt.want {
				t.Errorf("getConfigFromDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockLog struct {
	logs.Logger
	t *testing.T
}

func (l *MockLog) Errorf(template string, args ...interface{}) {
	l.t.Errorf(template, args...)
}

type MockTranslator struct {
	i18n.Translator
}

func (t *MockTranslator) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

func Test_provider_monitorConfigMap(t *testing.T) {
	type args struct {
		ns string
	}
	os.Unsetenv("METRIC_INDEX_TTL")
	os.Unsetenv("LOG_TTL")
	tests := []struct {
		name string
		args args
		want map[string]*pb.ConfigItem
	}{
		{
			name: "prod test",
			args: args{
				ns: "prod",
			},
			want: map[string]*pb.ConfigItem{
				"logs_ttl": {
					Key:   "logs_ttl",
					Name:  "logs_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(7)),
					Unit:  "days",
				},
				"logs_hot_ttl": {
					Key:   "logs_hot_ttl",
					Name:  "logs_hot_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(3)),
					Unit:  "days",
				},
				"metrics_ttl": {
					Key:   "metrics_ttl",
					Name:  "metrics_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(7)),
					Unit:  "days",
				},
				"metrics_hot_ttl": {
					Key:   "metrics_hot_ttl",
					Name:  "metrics_hot_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(3)),
					Unit:  "days",
				},
			},
		},
		{
			name: "default test",
			args: args{
				ns: "default",
			},
			want: map[string]*pb.ConfigItem{
				"logs_ttl": {
					Key:   "logs_ttl",
					Name:  "logs_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(3)),
					Unit:  "days",
				},
				"logs_hot_ttl": {
					Key:   "logs_hot_ttl",
					Name:  "logs_hot_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(1)),
					Unit:  "days",
				},
				"metrics_ttl": {
					Key:   "metrics_ttl",
					Name:  "metrics_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(3)),
					Unit:  "days",
				},
				"metrics_hot_ttl": {
					Key:   "metrics_hot_ttl",
					Name:  "metrics_hot_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(1)),
					Unit:  "days",
				},
			},
		},
		{
			name: "general",
			args: args{
				ns: "general",
			},
			want: map[string]*pb.ConfigItem{
				"metrics_ttl": &pb.ConfigItem{
					Key:   "metrics_ttl",
					Name:  "base metrics_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(3)),
					Unit:  "days",
				},
				"metrics_hot_ttl": &pb.ConfigItem{
					Key:   "metrics_hot_ttl",
					Name:  "base metrics_hot_ttl",
					Type:  "number",
					Value: structpb.NewNumberValue(float64(1)),
					Unit:  "days",
				},
			},
		},
	}
	ss := &settingsService{
		p: &provider{
			Log: &MockLog{
				t: t,
			},
		},
		t: &MockTranslator{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDefine := ss.monitorConfigMap(tt.args.ns)
			require.NotNil(t, configDefine)
			require.NotNil(t, configDefine.defaults)

			t.Log(configDefine.defaults, "defaults!")
			for k, v := range tt.want {
				require.EqualValuesf(t, v, configDefine.defaults[k](i18n.LanguageCodes{}), "key:%s", k)
			}
		})
	}
}
