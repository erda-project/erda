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

package orgapis

import (
	"testing"
)

func Test_calculateStatus(t *testing.T) {
	type args struct {
		raws []rawStatus
		name string
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{"status-0", args{
			raws: []rawStatus{
				{
					HealthStatus: 0,
					Weight:       0,
				},
				{
					HealthStatus: 0,
					Weight:       1,
				},
			},
			name: "kubernetes",
		}, 0},
		{"status-1", args{
			raws: []rawStatus{
				{
					HealthStatus: 1,
					Weight:       0,
				},
				{
					HealthStatus: 1,
					Weight:       1,
				},
			},
			name: "kubernetes",
		}, 1},
		{"status-2", args{
			raws: []rawStatus{
				{
					HealthStatus: 0,
					Weight:       0,
				},
				{
					HealthStatus: 1,
					Weight:       0,
				},
				{
					HealthStatus: 2,
					Weight:       0,
				},
			},
			name: "kubernetes",
		}, 1},
		{"status-2", args{
			raws: []rawStatus{
				{
					HealthStatus: 0,
					Weight:       0,
				},
				{
					HealthStatus: 1,
					Weight:       1,
				},
				{
					HealthStatus: 2,
					Weight:       1,
				},
			},
			name: "kubernetes",
		}, 2},
		{"status-3", args{
			raws: []rawStatus{
				{
					HealthStatus: 0,
					Weight:       0,
				},
				{
					HealthStatus: 1,
					Weight:       0,
				},
				{
					HealthStatus: 2,
					Weight:       1,
				},
			},
			name: "machine",
		}, 3},
		{"status-3", args{
			raws: []rawStatus{
				{
					HealthStatus: 2,
					Weight:       0,
				},
				{
					HealthStatus: 1,
					Weight:       0,
				},
				{
					HealthStatus: 2,
					Weight:       0,
				},
			},
			name: "machine",
		}, 2},
		{"status-3", args{
			raws: []rawStatus{
				{
					HealthStatus: 2,
					Weight:       1,
				},
				{
					HealthStatus: 2,
					Weight:       1,
				},
				{
					HealthStatus: 2,
					Weight:       1,
				},
			},
			name: "kubernetes",
		}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStatus(tt.args.raws, tt.args.name); got != tt.want {
				t.Errorf("calculateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
