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

package db

import (
	"reflect"
	"testing"
)

func TestFilterByEvent(t *testing.T) {
	type args struct {
		triggers []PipelineTrigger
		Filter   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []PipelineTrigger
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				triggers: []PipelineTrigger{
					{
						Event: "ssss",
						Filter: map[string]string{
							"appID":     "ss",
							"projectID": "xx",
						},
					},
					{
						Event: "ssss",
						Filter: map[string]string{
							"appID":     "ss",
							"projectID": "xx",
							"branch":    "master",
						},
					},
					{
						Event: "ssss",
						Filter: map[string]string{
							"appID":     "ss",
							"projectID": "xx",
							"branch":    "master",
							"orgID":     "1",
						},
					},
				},
				Filter: map[string]string{
					"appID":     "ss",
					"projectID": "xx",
					"branch":    "master",
					"orgID":     "1",
				},
			},
			want: []PipelineTrigger{
				{
					Event: "ssss",
					Filter: map[string]string{
						"appID":     "ss",
						"projectID": "xx",
					},
				},
				{
					Event: "ssss",
					Filter: map[string]string{
						"appID":     "ss",
						"projectID": "xx",
						"branch":    "master",
					},
				},
				{
					Event: "ssss",
					Filter: map[string]string{
						"appID":     "ss",
						"projectID": "xx",
						"branch":    "master",
						"orgID":     "1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterByEvent(tt.args.triggers, tt.args.Filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterByEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterByEvent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
