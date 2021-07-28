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

package storage

import (
	"bytes"
	"compress/gzip"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	logmodule "github.com/erda-project/erda/modules/core/monitor/log"
)

func BenchmarkGzipContentV1(b *testing.B) {
	s := strings.Builder{}
	for i := 0; i < 1000; i++ {
		s.WriteString("*")
	}
	b.ReportAllocs()

	content := s.String()
	for i := 0; i < b.N; i++ {
		gzipContent(content)
	}
}

func BenchmarkGzipContentV2(b *testing.B) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	s := strings.Builder{}
	for i := 0; i < 1000; i++ {
		s.WriteString("*")
	}
	b.ReportAllocs()

	content := s.String()
	for i := 0; i < b.N; i++ {
		gzipContentV2(content, w)
	}
}

func TestLogStatement_GetStatement(t *testing.T) {
	var stringPtr *string

	type fields struct {
		p *provider
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		{
			name:   "logs.LogMeta",
			fields: fields{p: mockProvider()},
			args: args{data: &logmodule.LogMeta{
				ID:     "aaa",
				Source: "container",
				Tags:   map[string]string{"level": "INFO"},
			}},
			want:    "INSERT INTO spot_prod.base_log_meta (source, id, tags) VALUES (?, ?, ?) USING TTL ?;",
			want1:   []interface{}{"container", "aaa", map[string]string{"level": "INFO"}, 60},
			wantErr: false,
		},
		{
			name:   "pb.Log",
			fields: fields{p: mockProvider()},
			args: args{data: &logmodule.Log{
				ID:        "aaa",
				Source:    "container",
				Stream:    "stdout",
				Offset:    1024,
				Timestamp: 1604892459000000000,
				Content:   "hello world",
				Tags:      map[string]string{"level": "INFO", "dice_org_name": "org1"},
			}},
			want: "INSERT INTO spot_org1.base_log (source, id, stream, time_bucket, timestamp, offset, content, level, request_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) USING TTL ?;",
			want1: []interface{}{
				"container",
				"aaa",
				"stdout",
				int64(1604880000000000000),
				int64(1604892459000000000),
				int64(1024),
				gzipString("hello world"),
				"INFO",
				stringPtr,
				60,
			},
			wantErr: false,
		},
		{
			name:   "pb.Log with request-id",
			fields: fields{p: mockProvider()},
			args: args{data: &logmodule.Log{
				ID:        "aaa",
				Source:    "container",
				Stream:    "stdout",
				Offset:    1024,
				Timestamp: 1604892459000000000,
				Content:   "hello world",
				Tags:      map[string]string{"level": "INFO", "dice_org_name": "org1", "request-id": "bbb"},
			}},
			want: "INSERT INTO spot_org1.base_log (source, id, stream, time_bucket, timestamp, offset, content, level, request_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) USING TTL ?;",
			want1: []interface{}{
				"container",
				"aaa",
				"stdout",
				int64(1604880000000000000),
				int64(1604892459000000000),
				int64(1024),
				gzipString("hello world"),
				"INFO",
				ptrString("bbb"),
				60,
			},
			wantErr: false,
		},
		{
			name:    "bad type",
			fields:  fields{p: mockProvider()},
			args:    args{data: "hello"},
			want:    "",
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := tt.fields.p.createLogStatementBuilder()
			got, got1, err := ls.GetStatement(tt.args.data)
			ass := assert.New(t)
			ass.Equal(tt.wantErr, err != nil)
			ass.Equal(tt.want, got)
			ass.Equal(tt.want1, got1)
		})
	}
}

func gzipString(data string) []byte {
	d, err := gzipContent(data)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func ptrString(s string) *string {
	return &s
}
