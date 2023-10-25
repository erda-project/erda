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

package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendCommonHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name:    "empty",
			headers: map[string]string{"hello": "world"},
			expected: map[string]string{
				"hello":         "world",
				"Cache-Control": "no-cache",
				"Pragma":        "no-cache",
				"Connection":    "keep-alive",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendCommonHeaders(tt.headers)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendCommonHeaders(), got: %v, want: %v", got, tt.expected)
			}
		})
	}
}

func TestEDASAppInfo(t *testing.T) {
	type args struct {
		sgType      string
		sgID        string
		serviceName string
	}

	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{
			name: "compose app info",
			args: args{
				sgType:      "service",
				sgID:        "1",
				serviceName: "app-demo",
			},
			expect: "service-1-app-demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := CombineEDASAppInfo(tt.args.sgType, tt.args.sgID, tt.args.serviceName); got != tt.expect {
				t.Fatalf("CombineEDASAppInfo, expecte: %v, got: %v", tt.expect, got)
			}
		})
	}
}

func TestCombineEDASAppGroup(t *testing.T) {
	group := CombineEDASAppGroup("type", "id")
	assert.Equal(t, "type-id", group)
}

func TestCombineEDASAppNameWithGroup(t *testing.T) {
	appName := CombineEDASAppNameWithGroup("group", "service")
	assert.Equal(t, "group-service", appName)
}

func TestCombineEDASAppName(t *testing.T) {
	appName := CombineEDASAppName("type", "id", "service")
	assert.Equal(t, "type-id-service", appName)
}

func TestCombineEDASAppInfo(t *testing.T) {
	group, appName := CombineEDASAppInfo("type", "id", "service")
	assert.Equal(t, "type-id", group)
	assert.Equal(t, "type-id-service", appName)
}

func TestMakeEnvVariableName(t *testing.T) {
	envName := MakeEnvVariableName("my-environment-variable")
	assert.Equal(t, "MY_ENVIRONMENT_VARIABLE", envName)
}

func TestEnvToString(t *testing.T) {
	tests := []struct {
		name     string
		envs     map[string]string
		expected string
		wantErr  bool
	}{
		{
			name: "case1",
			envs: map[string]string{
				"KEY1": "VALUE1",
				"KEY2": "VALUE2",
			},
			expected: `[{"name":"KEY1","value":"VALUE1"},{"name":"KEY2","value":"VALUE2"}]`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := EnvToString(tt.envs); (err != nil) != tt.wantErr {
				t.Errorf("EnvToString() error = %v, wantErr %v", err, tt.wantErr)
			} else if got != tt.expected {
				t.Errorf("EnvToString() expect: %s, got: %s", tt.expected, got)
			}
		})
	}
}
