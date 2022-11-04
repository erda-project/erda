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
	"fmt"
	"os"
	"testing"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
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
		days int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_getConfigFromDays",
			args: args{
				3,
			},
			want: `{"ttl":"72h0m0s"}`,
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
	return ""
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
		want *configDefine
	}{
		{
			name: "test_provider_monitorConfigMap",
			args: args{
				ns: "prod",
			},
			want: nil,
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
			if got := ss.monitorConfigMap(tt.args.ns); got != nil {
				fmt.Printf("monitorConfigMap() = %+v", got)
			}
		})
	}
}

// func Test_provider_syncMonitorConfig(t *testing.T) {
// 	p := &provider{
// 		L: &pjyLog{
// 			Entry: &logrus.Entry{
// 				Logger: &logrus.Logger{},
// 			},
// 		},
// 	}
// 	db := &gorm.DB{}
// 	exec := monkey.PatchInstanceMethod(reflect.TypeOf(db), "Exec", func(_ *gorm.DB,
// 		sql string, values ...interface{}) *gorm.DB {
// 		return &gorm.DB{
// 			Error: nil,
// 		}
// 	})
// 	defer exec.Unpatch()
// 	orgid := 1
// 	orgName := "terminus"
// 	list := []*monitorConfigRegister{
// 		{
// 			ScopeID: "18",
// 			Filters: `[{"key":"erda","value":"pjy"}]`,
// 			Names:   "terminus",
// 		},
// 	}
// 	err := p.syncMonitorConfig(db, orgid, "1", orgName, "local", "metric", "erda", list, 3)
// 	assert.Equal(t, nil, err)
// }

// func Test_insertOrgFilter1(t *testing.T) {
// 	type args struct {
// 		typ     string
// 		orgID   string
// 		orgName string
// 		filters string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		{
// 			name: "test_insertOrgFilter1",
// 			args: args{
// 				typ:     "metric",
// 				orgID:   "",
// 				orgName: "terminus",
// 				filters: `[{"key":"erda","value":"pjy"}]`,
// 			},
// 		},
// 		{
// 			name: "test_insertOrgFilter1",
// 			args: args{
// 				typ:     "log",
// 				orgID:   "",
// 				orgName: "terminus",
// 				filters: `[{"key":"erda","value":"pjy"}]`,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := insertOrgFilter(tt.args.typ, tt.args.orgID, tt.args.orgName, tt.args.filters)
// 			if err != nil {
// 				t.Errorf("insertOrgFilter() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			fmt.Printf("insertOrgFilter() got = %+v", got)
// 		})
// 	}
// }

// func Test_getConfigFromDays1(t *testing.T) {
// 	type args struct {
// 		days int64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{
// 			name: "test_getConfigFromDays1",
// 			args: args{
// 				days: 3,
// 			},
// 			want: `{"ttl":"72h0m0s"}`,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getConfigFromDays(tt.args.days); got != tt.want {
// 				t.Errorf("getConfigFromDays() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
