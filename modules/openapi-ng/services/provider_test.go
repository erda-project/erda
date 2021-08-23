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

package services

import (
	"reflect"
	"testing"
)

func Test_buildPathToSegments(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSegs []*pathSegment
	}{
		{
			path: "/abc/def",
			wantSegs: []*pathSegment{
				{
					typ:  pathStatic,
					name: "/abc/def",
				},
			},
		},
		{
			path: "{def}",
			wantSegs: []*pathSegment{
				{
					typ:  pathField,
					name: "def",
				},
			},
		},
		{
			path: "/abc/{def}",
			wantSegs: []*pathSegment{
				{
					typ:  pathStatic,
					name: "/abc/",
				},
				{
					typ:  pathField,
					name: "def",
				},
			},
		},
		{
			path: "{abc}/def",
			wantSegs: []*pathSegment{
				{
					typ:  pathField,
					name: "abc",
				},
				{
					typ:  pathStatic,
					name: "/def",
				},
			},
		},
		{
			path: "/abc/{def}/g",
			wantSegs: []*pathSegment{
				{
					typ:  pathStatic,
					name: "/abc/",
				},
				{
					typ:  pathField,
					name: "def",
				},
				{
					typ:  pathStatic,
					name: "/g",
				},
			},
		},
		{
			path: "/abc/{def=subpath/**}/g",
			wantSegs: []*pathSegment{
				{
					typ:  pathStatic,
					name: "/abc/",
				},
				{
					typ:  pathField,
					name: "def",
				},
				{
					typ:  pathStatic,
					name: "/g",
				},
			},
		},
		{
			path: "/abc/{def=subpath/**}",
			wantSegs: []*pathSegment{
				{
					typ:  pathStatic,
					name: "/abc/",
				},
				{
					typ:  pathField,
					name: "def",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSegs := buildPathToSegments(tt.path); !reflect.DeepEqual(gotSegs, tt.wantSegs) {
				t.Errorf("buildPathToSegments() = %v, want %v", gotSegs, tt.wantSegs)
			}
		})
	}
}
