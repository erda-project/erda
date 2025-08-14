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

package handler_mcp_server

import (
	"testing"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/mcp_server/pb"
)

func Test_buildEndpoint(t *testing.T) {

	handler := &MCPHandler{
		McpProxyPublicURL: "http://127.0.0.1:8081",
	}

	server := &pb.MCPServer{
		Name:    "demo",
		Version: "1.0.0",
	}

	got := handler.buildEndpoint(server)
	want := "http://127.0.0.1:8081/proxy/connect/demo/1.0.0"
	if got != want {
		t.Errorf("buildEndpoint() got = %q, want %q", got, want)
	}
}

func Test_VerifyAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			addr:    "not_a_url",
			wantErr: true,
		},
		{
			name:    "invalid port non-numeric",
			addr:    "127.0.0.1:abc",
			wantErr: true,
		},
		{
			name:    "invalid port out of range",
			addr:    "127.0.0.1:70000",
			wantErr: true,
		},
		{
			name:    "valid without port",
			addr:    "http://127.0.0.1",
			wantErr: false,
		},
		{
			name:    "valid with port",
			addr:    "http://127.0.0.1:8080",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyAddr(tt.addr)
			if !tt.wantErr {
				if err != nil {
					t.Errorf("VerifyAddr(%q) unexpected error: %v", tt.addr, err)
				}
			} else {
				if err == nil {
					t.Errorf("VerifyAddr(%q) got error %v, want %v", tt.addr, err, tt.wantErr)
				}
			}
		})
	}
}
