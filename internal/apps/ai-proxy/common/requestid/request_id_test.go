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

package requestid

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

// UUID format regex: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestGetOrSetRequestID(t *testing.T) {
	tests := []struct {
		name              string
		existingRequestID string
		expectNewUUID     bool
	}{
		{
			name:              "existing request ID",
			existingRequestID: "existing-request-123",
			expectNewUUID:     false,
		},
		{
			name:              "no existing request ID",
			existingRequestID: "",
			expectNewUUID:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set(vars.XRequestId, tt.existingRequestID)
			}

			result := GetOrGenRequestID(req)

			if tt.expectNewUUID {
				// Should generate a new UUID
				assert.True(t, uuidRegex.MatchString(result), "Result should be a valid UUID")
			} else {
				// Should return the existing value
				assert.Equal(t, tt.existingRequestID, result)
				assert.Equal(t, tt.existingRequestID, req.Header.Get(vars.XRequestId))
			}
		})
	}
}

func TestGetCallID(t *testing.T) {
	tests := []struct {
		name              string
		existingRequestID string
	}{
		{
			name:              "with existing request ID",
			existingRequestID: "existing-request-123",
		},
		{
			name:              "without existing request ID",
			existingRequestID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set(vars.XRequestId, tt.existingRequestID)
			}

			originalHeaderValue := req.Header.Get(vars.XRequestId)
			result := GetCallID(req)

			// Should always generate a new UUID
			assert.True(t, uuidRegex.MatchString(result), "Result should be a valid UUID")

			// Should not modify the request header
			assert.Equal(t, originalHeaderValue, req.Header.Get(vars.XRequestId))
		})
	}
}

func TestGetCallID_AlwaysGeneratesNewUUID(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	// Generate multiple call IDs
	callID1 := GetCallID(req)
	callID2 := GetCallID(req)
	callID3 := GetCallID(req)

	// All should be valid UUIDs
	assert.True(t, uuidRegex.MatchString(callID1))
	assert.True(t, uuidRegex.MatchString(callID2))
	assert.True(t, uuidRegex.MatchString(callID3))

	// All should be different
	assert.NotEqual(t, callID1, callID2)
	assert.NotEqual(t, callID2, callID3)
	assert.NotEqual(t, callID1, callID3)
}

func TestGetOrSetID(t *testing.T) {
	tests := []struct {
		name            string
		headerKey       string
		existingValue   string
		expectSetHeader bool
		expectNewUUID   bool
	}{
		{
			name:            "empty header key - should not set header",
			headerKey:       "",
			existingValue:   "",
			expectSetHeader: false,
			expectNewUUID:   true,
		},
		{
			name:            "with header key and existing value",
			headerKey:       "X-Test-Id",
			existingValue:   "existing-123",
			expectSetHeader: false,
			expectNewUUID:   false,
		},
		{
			name:            "with header key and no existing value",
			headerKey:       "X-Test-Id",
			existingValue:   "",
			expectSetHeader: false,
			expectNewUUID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingValue != "" {
				req.Header.Set(tt.headerKey, tt.existingValue)
			}

			result := getOrGenID(req, tt.headerKey)

			if tt.expectNewUUID {
				assert.True(t, uuidRegex.MatchString(result), "Result should be a valid UUID")
			} else {
				assert.Equal(t, tt.existingValue, result)
			}

			if tt.headerKey != "" && tt.existingValue != "" {
				assert.Equal(t, tt.existingValue, req.Header.Get(tt.headerKey), "Header should remain unchanged")
			}
		})
	}
}

func TestGetOrSetRequestID_Integration(t *testing.T) {
	// Test the integration with actual vars.XRequestId
	req := httptest.NewRequest("GET", "/test", nil)

	// First call should generate UUID but not set header
	result1 := GetOrGenRequestID(req)
	assert.True(t, uuidRegex.MatchString(result1))
	assert.Equal(t, "", req.Header.Get(vars.XRequestId))

	// Second call should generate a different UUID
	result2 := GetOrGenRequestID(req)
	assert.True(t, uuidRegex.MatchString(result2))
	assert.NotEqual(t, result1, result2)
	assert.Equal(t, "", req.Header.Get(vars.XRequestId))
}

func TestUUIDFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	// Test GetOrGenRequestID UUID format
	requestID := GetOrGenRequestID(req)
	assert.True(t, uuidRegex.MatchString(requestID), "GetOrGenRequestID should return valid UUID format")

	// Test GetCallID UUID format
	callID := GetCallID(req)
	assert.True(t, uuidRegex.MatchString(callID), "GetCallID should return valid UUID format")

	// Test that UUIDs are different
	assert.NotEqual(t, requestID, callID, "Request ID and Call ID should be different")
}
