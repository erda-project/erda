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
