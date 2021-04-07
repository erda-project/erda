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

package tsql

import (
	"testing"
	"time"
)

func Test_formatDuration(t *testing.T) {
	type args struct {
		d         time.Duration
		precision int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{

		{
			"",
			args{
				1*time.Microsecond + 999*time.Nanosecond,
				2,
			},
			"2µs",
			false,
		},
		{
			"",
			args{
				1*time.Microsecond + 999*time.Nanosecond,
				0,
			},
			"2µs",
			false,
		},
		{
			"",
			args{
				1*time.Millisecond + 999*time.Microsecond,
				0,
			},
			"2ms",
			false,
		},
		{
			"",
			args{
				1*time.Second + 999*time.Millisecond,
				0,
			},
			"2s",
			false,
		},
		{
			"",
			args{
				567*time.Microsecond + 700*time.Nanosecond,
				0,
			},
			"568µs",
			false,
		},
		{
			"",
			args{
				569*time.Microsecond + 700*time.Nanosecond,
				0,
			},
			"570µs",
			false,
		},
		{
			"",
			args{
				1*time.Second + 567*time.Millisecond,
				2,
			},
			"1.57s",
			false,
		},
		{
			"",
			args{
				1*time.Second + 100*time.Millisecond,
				5,
			},
			"1.1s",
			false,
		},
		{
			"",
			args{
				1 * time.Second,
				1,
			},
			"1s",
			false,
		},
		{
			"",
			args{
				1*time.Second + 100*time.Millisecond,
				1,
			},
			"1.1s",
			false,
		},
		{
			"",
			args{
				1*time.Second + 500*time.Millisecond,
				1,
			},
			"1.5s",
			false,
		},
		{
			"",
			args{
				30*time.Second + 100*time.Millisecond + 1*time.Microsecond,
				2,
			},
			"30.1s",
			false,
		},
		{
			"",
			args{
				30*time.Second + 115*time.Millisecond + 1*time.Microsecond,
				2,
			},
			"30.12s",
			false,
		},
		{
			"",
			args{
				24*time.Hour + 30*time.Second + 115*time.Millisecond + 1*time.Microsecond,
				2,
			},
			"24h0m30.12s",
			false,
		},
		{
			"",
			args{
				24*time.Hour + 30*time.Second + 115*time.Millisecond + 1*time.Microsecond,
				0,
			},
			"24h0m30s",
			false,
		},
		{
			"",
			args{
				10000*time.Hour + 59*time.Second + 99*time.Millisecond + 999*time.Microsecond + 999*time.Nanosecond,
				3,
			},
			"10000h0m59.1s",
			false,
		},
		{
			"",
			args{
				10000*time.Hour + 59*time.Second + 999*time.Millisecond + 999*time.Microsecond + 999*time.Nanosecond,
				10000,
			},
			"10000h0m59.999999999s",
			false,
		},
		{
			"",
			args{
				10000*time.Hour + 59*time.Second + 999*time.Millisecond + 999*time.Microsecond + 999*time.Nanosecond,
				8,
			},
			"10000h1m0s",
			false,
		},
		{
			"",
			args{
				24*time.Hour + 30*time.Second + 115*time.Millisecond + 1*time.Microsecond,
				-1,
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatDuration(tt.args.d, tt.args.precision)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
