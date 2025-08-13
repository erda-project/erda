package mcp_server_response

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
