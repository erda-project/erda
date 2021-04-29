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

package report

import (
	"bytes"
	"io"
	"testing"
)

func TestCompressWithGzip(t *testing.T) {
	type args struct {
		data io.Reader
	}
	r := bytes.NewReader([]byte{97, 98, 99, 100})
	tests := []struct {
		name    string
		args    args
		want    io.Reader
		wantErr bool
	}{
		{
			name: "test_CompressWithGzip",
			args: args{
				data: r,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompressWithGzip(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompressWithGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("CompressWithGzip() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
