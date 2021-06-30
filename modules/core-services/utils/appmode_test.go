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

package utils

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestCheckAppMode(t *testing.T) {
	type args struct {
		mode string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "otherName",
			args: args{
				mode: "_&((()))",
			},
			wantErr: true,
		},
		{
			name: string(apistructs.ApplicationModeService),
			args: args{
				mode: string(apistructs.ApplicationModeService),
			},
			wantErr: false,
		},
		{
			name: string(apistructs.ApplicationModeLibrary),
			args: args{
				mode: string(apistructs.ApplicationModeLibrary),
			},
			wantErr: false,
		},
		{
			name: string(apistructs.ApplicationModeMobile),
			args: args{
				mode: string(apistructs.ApplicationModeMobile),
			},
			wantErr: false,
		},
		{
			name: string(apistructs.ApplicationModeProjectService),
			args: args{
				mode: string(apistructs.ApplicationModeProjectService),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckAppMode(tt.args.mode); (err != nil) != tt.wantErr {
				t.Errorf("CheckAppMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
