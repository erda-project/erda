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

package base

import (
	"testing"
)

func TestGetScenarioAndCompNameFromProviderKey(t *testing.T) {
	type args struct {
		providerKey string
	}
	tests := []struct {
		name             string
		args             args
		wantScenario     string
		wantCompName     string
		wantInstanceName string
		haveErr          bool
	}{
		{
			name: "issue-manage.content",
			args: args{
				providerKey: "component-protocol.components.issue-manage.content",
			},
			wantScenario:     "issue-manage",
			wantCompName:     "content",
			wantInstanceName: "content",
			haveErr:          false,
		},
		{
			name: "issue-manage.content@content2",
			args: args{
				providerKey: "component-protocol.components.issue-manage.content@content2",
			},
			wantScenario:     "issue-manage",
			wantCompName:     "content",
			wantInstanceName: "content2",
			haveErr:          false,
		},
		{
			name: "missing compName",
			args: args{
				providerKey: "component-protocol.components.issue-manage",
			},
			wantScenario:     "",
			wantCompName:     "",
			wantInstanceName: "",
			haveErr:          true,
		},
		{
			name: "invalid prefix",
			args: args{
				providerKey: "xxx.components.issue-manage",
			},
			wantScenario:     "",
			wantCompName:     "",
			wantInstanceName: "",
			haveErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScenario, gotCompName, gotInstanceName, gotErr := GetScenarioAndCompNameFromProviderKey(tt.args.providerKey)
			if gotScenario != tt.wantScenario {
				t.Errorf("MustGetScenarioAndCompNameFromProviderKey() gotScenario = %v, want %v", gotScenario, tt.wantScenario)
			}
			if gotCompName != tt.wantCompName {
				t.Errorf("MustGetScenarioAndCompNameFromProviderKey() gotCompName = %v, want %v", gotCompName, tt.wantCompName)
			}
			if gotInstanceName != tt.wantInstanceName {
				t.Errorf("MustGetScenarioAndInstanceNameFromProviderKey() gotInstanceName = %v, want %v", gotInstanceName, tt.wantInstanceName)
			}
			if (tt.haveErr && gotErr == nil) || (!tt.haveErr && gotErr != nil) {
				t.Errorf("MustGetScenarioAndCompNameFromProviderKey() getErr = %v, haveErr %v", gotErr, tt.haveErr)
			}
		})
	}
}
