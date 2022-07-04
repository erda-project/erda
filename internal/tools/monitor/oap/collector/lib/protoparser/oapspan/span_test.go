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

package oapspan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
)

func Test_unmarshalWork_Unmarshal(t *testing.T) {
	type fields struct {
		buf []byte
		err error
	}

	tests := []struct {
		name    string
		fields  fields
		want    *trace.Span
		wantErr bool
	}{
		{
			fields: fields{
				buf: []byte(`{"traceID":"bbb","spanID":"aaa","parentSpanID":"","startTimeUnixNano":1652756014793553000,"endTimeUnixNano":1652756014793553000,"name":"GET /","relations":null,"attributes":{"hello":"world","org_name":"erda"}}`),
			},
			want: &trace.Span{
				SpanId:        "aaa",
				TraceId:       "bbb",
				ParentSpanId:  "",
				OperationName: "GET /",
				OrgName:       "erda",
				StartTime:     int64(1652756014793553000),
				EndTime:       int64(1652756014793553000),
				Tags: map[string]string{
					"hello":    "world",
					"org_name": "erda",
				},
			},
			wantErr: false,
		},
		{
			name: "parser error",
			fields: fields{
				buf: []byte(`"traceID":"bbb","spanID":"aaa","parentSpanID":"","startTimeUnixNano":1652756014793553000,"endTimeUnixNano":1652756014793553000,"name":"GET /","relations":null,"attributes":{"hello":"world","org_name":"erda"}}`),
			},
			want: &trace.Span{
				SpanId:        "aaa",
				TraceId:       "bbb",
				ParentSpanId:  "",
				OperationName: "GET /",
				OrgName:       "erda",
				StartTime:     int64(1652756014793553000),
				EndTime:       int64(1652756014793553000),
				Tags: map[string]string{
					"org_name": "erda",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uw := &unmarshalWork{
				buf: tt.fields.buf,
				err: tt.fields.err,
				callback: func(span *trace.Span) error {
					assert.Equal(t, tt.want, span)
					return nil
				},
			}
			uw.wg.Add(1)
			uw.Unmarshal()
			uw.wg.Wait()
			if !tt.wantErr {
				assert.Nil(t, uw.err)
			} else {
				assert.NotNil(t, uw.err)
			}
		})
	}
}
