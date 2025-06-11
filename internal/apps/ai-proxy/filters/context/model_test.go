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

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func Test_findModelID(t *testing.T) {
	type args struct {
		infor reverseproxy.HttpInfor
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "model id in header only",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(vars.XAIProxyModelId): []string{"model-id-in-header"},
					},
				}),
			},
			want:    "model-id-in-header",
			wantErr: false,
		},
		{
			name: "model id in body only",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey("Content-Type"): []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"model": "[ID:model-id-in-body]"}`)),
				}),
			},
			want:    "model-id-in-body",
			wantErr: false,
		},
		{
			name: "model id both in header and body, header is preferred",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey("Content-Type"):       []string{"application/json"},
						textproto.CanonicalMIMEHeaderKey(vars.XAIProxyModelId): []string{"model-id-in-header"},
					},
					Body: io.NopCloser(strings.NewReader(`{"model": "[ID:model-id-in-body]"}`)),
				}),
			},
			want:    "model-id-in-header",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findModelID(tt.args.infor)
			if (err != nil) != tt.wantErr {
				t.Errorf("findModelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findModelID() = %v, want %v", got, tt.want)
			}
		})
	}
}
