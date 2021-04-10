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

package adapt

import "testing"

func Test_convertDashboardURL(t *testing.T) {
	type args struct {
		domain      string
		path        string
		dashboardID string
		groups      []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_convertDashboardURL",
			args: args{
				domain:      "http",
				path:        "localhost:9090",
				dashboardID: "2",
				groups:      []string{"1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertDashboardURL(tt.args.domain, tt.args.path, tt.args.dashboardID, tt.args.groups); got != tt.want {
				t.Errorf("convertDashboardURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
