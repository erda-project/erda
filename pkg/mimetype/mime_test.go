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

package mimetype

import (
	"testing"
)

func TestTypeByFilename(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "xml",
			filePath: "test.xml",
			want:     "application/xml",
		},
		{
			name:     "json",
			filePath: "test.json",
			want:     "application/json",
		},
		{
			name:     "atom",
			filePath: "test.atom",
			want:     "application/atom+xml",
		},
		{
			name:     "path",
			filePath: "dir/test.json",
			want:     "application/json",
		},
		{
			name:     "unknown",
			filePath: "test.unknown",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeByFilename(tt.filePath); got != tt.want {
				t.Errorf("TypeByFilename(%q) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}
