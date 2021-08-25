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

package customhttp

import (
	"reflect"
	"testing"
)

func Test_parseInetUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name           string
		args           args
		wantPortalHost string
		wantPortalDest string
		wantPortalUrl  string
		wantPortalArgs map[string]string
		wantErr        bool
	}{
		// TODO: Add test cases.
		{
			"test1",
			args{"inet://abc/123"},
			"abc",
			"123",
			"",
			map[string]string{},
			false,
		},
		{
			"test2",
			args{"inet://abc"},
			"",
			"",
			"",
			map[string]string{},
			true,
		},
		{
			"test3",
			args{"inet://abc/123/qq?a=b"},
			"abc",
			"123",
			"qq?a=b",
			map[string]string{},
			false,
		},
		{
			"test4",
			args{"inet://abc?ssl=on&direct=on/123/qq?a=b"},
			"abc",
			"123",
			"qq?a=b",
			map[string]string{
				"ssl":    "on",
				"direct": "on",
			},
			false,
		},
		{
			"test5",
			args{"inet://abc?ssl=on&direct=on//123//qq?a=b"},
			"abc",
			"123",
			"qq?a=b",
			map[string]string{
				"ssl":    "on",
				"direct": "on",
			},
			false,
		},
		{
			"test6",
			args{"inet://abc?ssl=on&direct=on/123"},
			"abc",
			"123",
			"",
			map[string]string{
				"ssl":    "on",
				"direct": "on",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPortalHost, gotPortalDest, gotPortalUrl, gotPortalArgs, err := parseInetUrl(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInetUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPortalHost != tt.wantPortalHost {
				t.Errorf("parseInetUrl() gotPortalHost = %v, want %v", gotPortalHost, tt.wantPortalHost)
			}
			if gotPortalDest != tt.wantPortalDest {
				t.Errorf("parseInetUrl() gotPortalDest = %v, want %v", gotPortalDest, tt.wantPortalDest)
			}
			if gotPortalUrl != tt.wantPortalUrl {
				t.Errorf("parseInetUrl() gotPortalUrl = %v, want %v", gotPortalUrl, tt.wantPortalUrl)
			}
			if !reflect.DeepEqual(gotPortalArgs, tt.wantPortalArgs) {
				t.Errorf("parseInetUrl() gotPortalArgs = %v, want %v", gotPortalArgs, tt.wantPortalArgs)
			}
		})
	}
}
