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

package httputil_test

import (
	"net/http"
	"net/textproto"
	"testing"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func TestHeaderContains(t *testing.T) {
	var h = make(http.Header)
	h.Set("content-type", "application/json; charset=utf-8")
	ok := httputil.HeaderContains(h, httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains(h[textproto.CanonicalMIMEHeaderKey("content-type")], httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains([]httputil.ContentType{
		httputil.ApplicationJson,
		httputil.URLEncodedFormMime,
	}, httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains("application/json; charset=utf-8", httputil.ApplicationJson)
	t.Log(ok)
}

func TestContentTypeDetection(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		expectedMatch httputil.ContentType
		shouldMatch   bool
	}{
		{
			name:          "Exact match application/json",
			contentType:   "application/json",
			expectedMatch: httputil.ApplicationJson,
			shouldMatch:   true,
		},
		{
			name:          "Match application/json with charset",
			contentType:   "application/json; charset=utf-8",
			expectedMatch: httputil.ApplicationJson,
			shouldMatch:   true,
		},
		{
			name:          "Exact match application/json with charset",
			contentType:   "application/json; charset=utf-8",
			expectedMatch: httputil.ApplicationJsonUTF8,
			shouldMatch:   true,
		},
		{
			name:          "Case insensitive match",
			contentType:   "Application/JSON; Charset=UTF-8",
			expectedMatch: httputil.ApplicationJson,
			shouldMatch:   false, // HeaderContains is case sensitive
		},
		{
			name:          "Text event stream",
			contentType:   "text/event-stream",
			expectedMatch: httputil.TextEventStream,
			shouldMatch:   true,
		},
		{
			name:          "URL encoded form",
			contentType:   "application/x-www-form-urlencoded",
			expectedMatch: httputil.URLEncodedFormMime,
			shouldMatch:   true,
		},
		{
			name:          "No match",
			contentType:   "text/plain",
			expectedMatch: httputil.ApplicationJson,
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := make(http.Header)
			h.Set("Content-Type", tt.contentType)

			result := httputil.HeaderContains(h, tt.expectedMatch)
			if result != tt.shouldMatch {
				t.Errorf("Expected match %v, got %v for contentType '%s' against '%s'",
					tt.shouldMatch, result, tt.contentType, tt.expectedMatch)
			}

			// Also test direct string comparison
			stringResult := httputil.HeaderContains(tt.contentType, tt.expectedMatch)
			if stringResult != tt.shouldMatch {
				t.Errorf("String test: Expected match %v, got %v for contentType '%s' against '%s'",
					tt.shouldMatch, stringResult, tt.contentType, tt.expectedMatch)
			}
		})
	}
}
