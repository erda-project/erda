// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package settings

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/sirupsen/logrus"
)

type translator struct {
	common map[string]map[string]string
	dic    map[string]map[string]string
}

func (t *translator) Text(lang i18n.LanguageCodes, key string) string {
	return key
}

func (t *translator) Sprintf(lang i18n.LanguageCodes, key string, args ...interface{}) string {
	return key
}

func (t *translator) Get(lang i18n.LanguageCodes, key, def string) string {
	return def
}

func Test_provider_getDefaultConfig(t *testing.T) {
	type args struct {
		lang i18n.LanguageCodes
		ns   string
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   map[string]map[string]map[string]*configItem
	}{
		{
			name: "test_provider_getDefaultConfig",
			fields: provider{
				L: &pjyLog{
					Entry: &logrus.Entry{
						Logger: &logrus.Logger{},
					},
				},
				db:     nil,
				cfgMap: nil,
				t:      &translator{},
				bundle: nil,
			},
			args: args{
				lang: i18n.LanguageCodes{
					{
						Code: "zh",
					},
				},
				ns: "dev",
			},
			want: nil,
		},
	}
	tests[0].fields.initConfigMap()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				L:      tt.fields.L,
				db:     tt.fields.db,
				cfgMap: tt.fields.cfgMap,
				t:      tt.fields.t,
				bundle: tt.fields.bundle,
			}
			if got := p.getDefaultConfig(tt.args.lang, tt.args.ns); got != nil {
				fmt.Printf("getDefaultConfig() = %+v", got)
			}
		})
	}
}

func Test_getValue(t *testing.T) {
	type args struct {
		typ   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "test_getValue",
			args: args{
				typ:   "number",
				value: "3",
			},
			want: 3,
		},
		{
			name: "test_getValue",
			args: args{
				typ:   "string",
				value: "45",
			},
			want: "45",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getValue(tt.args.typ, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
