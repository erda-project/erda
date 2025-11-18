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

package dumplog

import (
	"testing"
)

func TestIsBinaryContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		expected    bool
		description string
	}{
		// Text types - should dump body
		{"text/plain", false, "plain text"},
		{"text/html", false, "HTML"},
		{"text/css", false, "CSS"},
		{"application/json", false, "JSON"},
		{"application/xml", false, "XML"},
		{"application/x-www-form-urlencoded", false, "form data"},
		{"application/javascript", false, "JavaScript"},
		{"text/plain; charset=utf-8", false, "text with charset"},
		{"application/json; charset=utf-8", false, "JSON with charset"},

		// Binary types - should not dump body
		{"audio/mpeg", true, "MP3 audio"},
		{"audio/wav", true, "WAV audio"},
		{"audio/mp4", true, "MP4 audio"},
		{"video/mp4", true, "MP4 video"},
		{"video/avi", true, "AVI video"},
		{"image/jpeg", true, "JPEG image"},
		{"image/png", true, "PNG image"},
		{"image/gif", true, "GIF image"},
		{"application/octet-stream", true, "binary stream"},
		{"application/pdf", true, "PDF document"},
		{"application/zip", true, "ZIP archive"},
		{"multipart/form-data", true, "multipart form"},
		{"multipart/form-data; boundary=something", true, "multipart form with boundary"},

		// Edge cases
		{"", false, "empty content type"},
		{"unknown/type", true, "unknown type - default to binary"},
		{"application/unknown", true, "unknown application type"},
		{"AUDIO/MPEG", true, "uppercase audio"},
		{"Audio/Mp3", true, "mixed case audio"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := IsBinaryContentType(tc.contentType)
			if result != tc.expected {
				t.Errorf("IsBinaryContentType(%q) = %v, expected %v",
					tc.contentType, result, tc.expected)
			}
		})
	}
}

func TestShouldDumpBody(t *testing.T) {
	testCases := []struct {
		contentType string
		expected    bool
		description string
	}{
		{"application/json", true, "JSON should be dumped"},
		{"text/plain", true, "text should be dumped"},
		{"audio/mpeg", false, "audio should not be dumped"},
		{"image/png", false, "image should not be dumped"},
		{"multipart/form-data", true, "multipart should be dumped"},
		{"", true, "empty content type should be dumped"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := ShouldDumpBody(tc.contentType)
			if result != tc.expected {
				t.Errorf("ShouldDumpBody(%q) = %v, expected %v",
					tc.contentType, result, tc.expected)
			}
		})
	}
}
