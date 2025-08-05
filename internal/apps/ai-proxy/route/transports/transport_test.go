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

package transports

import (
	"strings"
	"testing"
)

// Helper function: generate correctly formatted multipart data
func createMultipartBody(fields []MultipartField) string {
	var parts []string
	for _, field := range fields {
		if field.Filename != "" {
			// File field
			part := "---boundary123\r\n" +
				"Content-Disposition: form-data; name=\"" + field.Name + "\"; filename=\"" + field.Filename + "\"\r\n"
			if field.ContentType != "" {
				part += "Content-Type: " + field.ContentType + "\r\n"
			}
			part += "\r\n" + field.Value + "\r\n"
			parts = append(parts, part)
		} else {
			// Text field
			part := "---boundary123\r\n" +
				"Content-Disposition: form-data; name=\"" + field.Name + "\"\r\n" +
				"\r\n" + field.Value + "\r\n"
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "") + "---boundary123--\r\n"
}

type MultipartField struct {
	Name        string
	Value       string
	Filename    string
	ContentType string
}

func TestGenCurlPartsForMultipartForm(t *testing.T) {
	testCases := []struct {
		name           string
		initialCurl    string
		fields         []MultipartField
		expectedFields []string // --form parameters expected to appear in result
		description    string
	}{
		{
			name:        "Single text field",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "username", Value: "john_doe"},
			},
			expectedFields: []string{`-F username="john_doe"`},
			description:    "Single text field",
		},
		{
			name:        "Single file field",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "file", Value: "file content here", Filename: "test.txt", ContentType: "text/plain"},
			},
			expectedFields: []string{`-F file=@test.txt`},
			description:    "Single file field",
		},
		{
			name:        "Multiple text fields",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "model", Value: "whisper-1"},
				{Name: "language", Value: "en"},
			},
			expectedFields: []string{
				`-F model="whisper-1"`,
				`-F language="en"`,
			},
			description: "Multiple text fields",
		},
		{
			name:        "Mixed fields (text and file)",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "model", Value: "gpt-4"},
				{Name: "file", Value: "binary audio data here", Filename: "audio.mp3", ContentType: "audio/mpeg"},
				{Name: "temperature", Value: "0.7"},
			},
			expectedFields: []string{
				`-F model="gpt-4"`,
				`-F file=@audio.mp3`,
				`-F temperature="0.7"`,
			},
			description: "Mixed fields (text and file)",
		},
		{
			name:        "Audio transcription typical case",
			initialCurl: "curl -X POST 'https://api.openai.com/audio/transcriptions'",
			fields: []MultipartField{
				{Name: "file", Value: "fake binary audio data", Filename: "test_short.mp3", ContentType: "audio/mpeg"},
				{Name: "model", Value: "whisper-1"},
			},
			expectedFields: []string{
				`-F file=@test_short.mp3`,
				`-F model="whisper-1"`,
			},
			description: "Typical audio transcription request",
		},
		{
			name:        "Empty field value",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "optional_field", Value: ""},
				{Name: "required_field", Value: "value"},
			},
			expectedFields: []string{
				// Empty value fields are intentionally skipped, this is a reasonable design choice
				`-F required_field="value"`,
			},
			description: "Contains empty value field (empty value fields are skipped)",
		},
		{
			name:        "Field with multiline value",
			initialCurl: "curl -X POST",
			fields: []MultipartField{
				{Name: "description", Value: "This is a\nmultiline\ndescription"},
			},
			expectedFields: []string{
				`-F description="This is a\nmultiline\ndescription"`,
			},
			description: "Multiline text field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			multipartBody := createMultipartBody(tc.fields)
			result := genCurlPartsForMultipartForm(tc.initialCurl, []byte(multipartBody))

			// Verify result contains initial curl command
			if !strings.HasPrefix(result, tc.initialCurl) {
				t.Errorf("Result should start with initial curl command. Got: %s", result)
			}

			// Verify each expected field appears in result
			for _, expectedField := range tc.expectedFields {
				if !strings.Contains(result, expectedField) {
					t.Errorf("Expected field %s not found in result: %s", expectedField, result)
				}
			}

			// Verify the number of -F parameters is correct
			formCount := strings.Count(result, " -F ")
			expectedCount := len(tc.expectedFields)
			if formCount != expectedCount {
				t.Errorf("Expected %d -F parameters, but found %d in: %s",
					expectedCount, formCount, result)
			}

			t.Logf("✓ %s: %s", tc.description, result)
		})
	}
}

func TestGenCurlPartsForMultipartFormEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name          string
		initialCurl   string
		multipartBody string
		expectEmpty   bool
		description   string
	}{
		{
			name:          "Empty body",
			initialCurl:   "curl -X POST",
			multipartBody: "",
			expectEmpty:   true,
			description:   "Empty request body",
		},
		{
			name:        "No form fields",
			initialCurl: "curl -X POST",
			multipartBody: `------boundary123
------boundary123--`,
			expectEmpty: true,
			description: "No form fields",
		},
		{
			name:        "Malformed boundary",
			initialCurl: "curl -X POST",
			multipartBody: `malformed content without proper boundaries
Content-Disposition: form-data; name="field"
value`,
			expectEmpty: true,
			description: "Malformed boundary",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := genCurlPartsForMultipartForm(tc.initialCurl, []byte(tc.multipartBody))

			// Verify result contains initial curl command
			if !strings.HasPrefix(result, tc.initialCurl) {
				t.Errorf("Result should start with initial curl command. Got: %s", result)
			}

			if tc.expectEmpty {
				// For edge cases, should not add any -F parameters
				if strings.Contains(result, " -F ") {
					t.Errorf("Expected no -F parameters for edge case, but found some in: %s", result)
				}
			}

			t.Logf("✓ %s: %s", tc.description, result)
		})
	}
}

// Benchmark test
func BenchmarkGenCurlPartsForMultipartForm(b *testing.B) {
	initialCurl := "curl -X POST 'https://api.openai.com/audio/transcriptions'"
	multipartBody := `------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="file"; filename="test.mp3"
Content-Type: audio/mpeg

fake binary audio data that could be quite large in real scenarios
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="model"

whisper-1
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="language"

en
------WebKitFormBoundary7MA4YWxkTrZu0gW--`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genCurlPartsForMultipartForm(initialCurl, []byte(multipartBody))
	}
}
