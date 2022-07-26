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

package filesvc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_bytesToHexString(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "xss svg",
			path: "./example/xss-svg.svg",
			want: "3c3f786d6c2076657273696f6e3d22312e302220",
		},
		{
			name: "normal png",
			path: "./example/normal.png",
			want: "89504e470d0a1a0a0000000d4948445200000030",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.path)
			assert.NoError(t, err)
			defer f.Close()
			buf := make([]byte, 20)
			f.Read(buf)
			if got := bytesToHexString(buf); got != tt.want {
				t.Errorf("bytesToHexString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileContentType(t *testing.T) {
	tests := []struct {
		name string
		path string
		ext  string
		want string
	}{
		{
			name: "xss svg",
			path: "./example/xss-svg.svg",
			ext:  ".svg",
			want: "application/octet-stream",
		},
		{
			name: "normal png",
			path: "./example/normal.png",
			ext:  ".png",
			want: headerContentTypePng,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.path)
			assert.NoError(t, err)
			defer f.Close()
			if got := GetFileContentType(f, tt.ext); got != tt.want {
				t.Errorf("bytesToHexString() = %v, want %v", got, tt.want)
			}
		})
	}
}
