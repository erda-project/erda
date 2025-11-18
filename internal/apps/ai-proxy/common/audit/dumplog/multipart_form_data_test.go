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
	"bytes"
	"fmt"
	"mime/multipart"
	"strings"
	"testing"
)

func buildMultipartBody(t *testing.T) ([]byte, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("model", "openai/gpt-image-1 - test"); err != nil {
		t.Fatalf("failed to write model field: %v", err)
	}
	if err := writer.WriteField("prompt", "把毛毯改成红色的"); err != nil {
		t.Fatalf("failed to write prompt field: %v", err)
	}
	imagePart, err := writer.CreateFormFile("image[]", "cat.png")
	if err != nil {
		t.Fatalf("failed to create image part: %v", err)
	}
	if _, err := imagePart.Write([]byte{0x01, 0x02, 0x03}); err != nil {
		t.Fatalf("failed to write image part: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	return body.Bytes(), writer.FormDataContentType()
}

func TestFormatMultipartFormData(t *testing.T) {
	body, contentType := buildMultipartBody(t)

	formatted, err := formatMultipartFormData(contentType, body)
	if err != nil {
		t.Fatalf("formatMultipartFormData error: %v", err)
	}

	want := strings.Join([]string{
		"Form Data:",
		"model: openai/gpt-image-1 - test",
		"prompt: 把毛毯改成红色的",
		"image[]: (binary)",
	}, "\n")
	if formatted != want {
		t.Fatalf("unexpected formatted output:\nwant:\n%s\n\ngot:\n%s", want, formatted)
	}
}

func TestSanitizeMultipartDump(t *testing.T) {
	body, contentType := buildMultipartBody(t)
	header := fmt.Sprintf("POST /v1/images HTTP/1.1\r\nHost: example.com\r\nContent-Type: %s", contentType)

	dump := append([]byte(header+"\r\n\r\n"), body...)

	sanitized, err := sanitizeMultipartDump(dump, contentType)
	if err != nil {
		t.Fatalf("sanitizeMultipartDump error: %v", err)
	}

	expectedBody := strings.Join([]string{
		"Form Data:",
		"model: openai/gpt-image-1 - test",
		"prompt: 把毛毯改成红色的",
		"image[]: (binary)",
	}, "\n")
	want := header + "\r\n\r\n" + expectedBody
	if sanitized != want {
		t.Fatalf("unexpected sanitized dump:\nwant:\n%s\n\n got:\n%s", want, sanitized)
	}
}
