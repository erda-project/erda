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

package mcp_server_response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSessionId(t *testing.T) {
	tests := []struct {
		Message string
		Want    struct {
			Router    string
			SessionId string
		}
	}{
		{
			Message: "data: /message?sessionId=c4712579-9753-4fde-9129-92519abdd7c5\n",
			Want: struct {
				Router    string
				SessionId string
			}{Router: "/message", SessionId: "c4712579-9753-4fde-9129-92519abdd7c5"},
		}, {
			Message: "data: /proxy/message?sessionId=c4712579-9753-4fde-9129-92519abdd7c5\n",
			Want: struct {
				Router    string
				SessionId string
			}{Router: "/proxy/message", SessionId: "c4712579-9753-4fde-9129-92519abdd7c5"},
		}, {
			Message: "data: messages/?session_id=582c5949d3b54a2399a8366a646a6c81&ak=thisisaccesskey\n", // baidu mcp server
			Want: struct {
				Router    string
				SessionId string
			}{Router: "/messages/", SessionId: "582c5949d3b54a2399a8366a646a6c81"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Message, func(t *testing.T) {
			router, sessionId, err := parseSessionId(tt.Message)
			assert.NoError(t, err)
			assert.Equal(t, tt.Want.Router, router)
			assert.Equal(t, tt.Want.SessionId, sessionId)
		})
	}
}
