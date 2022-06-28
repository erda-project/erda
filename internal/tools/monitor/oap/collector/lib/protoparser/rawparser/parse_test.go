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

package rawparser

import (
	"bytes"
	"testing"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/compressor"
	"github.com/stretchr/testify/assert"
)

func TestParseStream(t *testing.T) {
	compress := compressor.NewGzipEncoder(9)
	type args struct {
		buf                []byte
		contentEncoding    string
		cusContentEncoding string
		format             string
		callback           func(buf []byte) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "json lines",
			args: args{
				buf:                []byte("{\"time\":\"2021-12-01T17:55:56.027178579+08:00\",\"stream\":\"stderr\",\"_p\":\"F\",\"log\":\"2\"}\n{\"time\":\"2021-12-01T17:55:56.027178579+08:00\",\"stream\":\"stderr\",\"_p\":\"F\",\"log\":\"3\"}\n"),
				contentEncoding:    "",
				cusContentEncoding: "",
				format:             "jsonl",
				callback: func(buf []byte) error {
					assert.Len(t, buf, 83)
					return nil
				},
			},
		},
		{
			name: "json array",
			args: args{
				buf:                []byte(`[{"time":"2021-12-01T17:55:56.027208973+08:00","stream":"stderr","log":"1"},{"time":"2021-12-01T17:55:56.027208973+08:00","stream":"stderr","log":"2"}]`),
				contentEncoding:    "gzip",
				cusContentEncoding: "",
				format:             "",
				callback: func(buf []byte) error {
					assert.Len(t, buf, 74)
					return nil
				},
			},
		},
		{
			name: "json line err",
			args: args{
				buf:                []byte(`[{"time":"2021-12-01T17:55:56.027208973+08:00","stream":"stderr","log":"1"},{"time":"2021-12-01T17:55:56.027208973+08:00","stream":"stderr","log":"2"}]`),
				contentEncoding:    "",
				cusContentEncoding: "",
				format:             "jsonl",
				callback: func(buf []byte) error {
					return nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf []byte
			switch tt.args.contentEncoding {
			case "gzip":
				gr, _ := compress.Compress(tt.args.buf)
				buf = gr
			default:
				buf = tt.args.buf
			}
			r := bytes.NewBuffer(buf)
			if err := ParseStream(r, tt.args.contentEncoding, tt.args.cusContentEncoding, tt.args.format, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("ParseStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
