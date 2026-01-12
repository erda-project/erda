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

package context

import "testing"

func Test_getMultipartBoundary(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name         string
		args         args
		wantBoundary string
		wantErr      bool
	}{
		{
			name: "normal",
			args: args{
				contentType: "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
			},
			wantBoundary: "----WebKitFormBoundary7MA4YWxkTrZu0gW",
			wantErr:      false,
		},
		{
			name: "no boundary",
			args: args{
				contentType: "multipart/form-data",
			},
			wantBoundary: "",
			wantErr:      true,
		},
		{
			name: "not multipart",
			args: args{
				contentType: "application/json",
			},
			wantBoundary: "",
			wantErr:      true,
		},
		{
			name: "empty",
			args: args{
				contentType: "",
			},
			wantBoundary: "",
			wantErr:      true,
		},
		{
			name: "charset before boundary",
			args: args{
				contentType: "multipart/form-data; charset=ISO-8859-1; boundary=a280vn9C0nIXdKhgKiepwuQnj8bM6rWUE49esIq",
			},
			wantBoundary: "a280vn9C0nIXdKhgKiepwuQnj8bM6rWUE49esIq",
			wantErr:      false,
		},
		{
			name: "quoted boundary",
			args: args{
				contentType: `multipart/form-data; boundary="----WebKitFormBoundaryABC123"`,
			},
			wantBoundary: "----WebKitFormBoundaryABC123",
			wantErr:      false,
		},
		{
			name: "mixed case multipart",
			args: args{
				contentType: "Multipart/Form-Data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
			},
			wantBoundary: "----WebKitFormBoundary7MA4YWxkTrZu0gW",
			wantErr:      false,
		},
		{
			name: "empty boundary param",
			args: args{
				contentType: "multipart/form-data; boundary=",
			},
			wantBoundary: "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBoundary, err := getMultipartBoundary(tt.args.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMultipartBoundary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBoundary != tt.wantBoundary {
				t.Errorf("getMultipartBoundary() gotBoundary = %v, want %v", gotBoundary, tt.wantBoundary)
			}
		})
	}
}
