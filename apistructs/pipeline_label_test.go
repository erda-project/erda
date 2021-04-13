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

package apistructs

import (
	"testing"
)

func TestPipelineLabelType_String(t *testing.T) {
	tests := []struct {
		name string
		t    PipelineLabelType
		want string
	}{
		{
			name: "queue",
			t:    PipelineLabelTypeQueue,
			want: "queue",
		},
		{
			name: "instance",
			t:    PipelineLabelTypeInstance,
			want: "p_i",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineLabelType_Valid(t *testing.T) {
	tests := []struct {
		name string
		t    PipelineLabelType
		want bool
	}{
		{
			name: "queue",
			t:    PipelineLabelTypeQueue,
			want: true,
		},
		{
			name: "instance",
			t:    PipelineLabelTypeInstance,
			want: true,
		},
		{
			name: "invalid",
			t:    "invalid",
			want: false,
		},
		{
			name: "string_queue",
			t:    "queue",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
