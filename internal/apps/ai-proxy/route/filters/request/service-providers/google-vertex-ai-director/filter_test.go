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

package google_vertex_ai_director

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestParseImageEditMultipart(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("prompt", "make it red"); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	imgPart, err := writer.CreateFormFile("image[]", "cat.png")
	if err != nil {
		t.Fatalf("create image part: %v", err)
	}
	if _, err := imgPart.Write([]byte{0x01, 0x02, 0x03}); err != nil {
		t.Fatalf("write image data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	payload, err := parseImageEditMultipart(bytes.NewReader(body.Bytes()), writer.FormDataContentType())
	if err != nil {
		t.Fatalf("parseImageEditMultipart returned error: %v", err)
	}

	if payload.Prompt != "make it red" {
		t.Fatalf("unexpected prompt: %q", payload.Prompt)
	}
	wantImage := []byte{0x01, 0x02, 0x03}
	if !bytes.Equal(payload.Image, wantImage) {
		t.Fatalf("unexpected image bytes: %v", payload.Image)
	}
	if payload.ImageContentType == "" {
		t.Fatalf("expected image content-type to be filled")
	}
}

func TestParseImageEditMultipart_InvalidContentType(t *testing.T) {
	if _, err := parseImageEditMultipart(bytes.NewReader(nil), "application/json"); err == nil {
		t.Fatalf("expected error for non-multipart content-type")
	}
}

func TestConvertImageSize(t *testing.T) {
	tests := []struct {
		name string
		req  openai.ImageRequest
		want string
	}{
		{name: "default", req: openai.ImageRequest{}, want: "1K"},
		{name: "large size", req: openai.ImageRequest{Size: openai.CreateImageSize1536x1024}, want: "2K"},
		{name: "hd quality", req: openai.ImageRequest{Size: openai.CreateImageSize1024x1024, Quality: openai.CreateImageQualityHD}, want: "2K"},
		{name: "high quality without size", req: openai.ImageRequest{Quality: openai.CreateImageQualityHigh}, want: "2K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertImageSize(tt.req); got != tt.want {
				t.Fatalf("convertImageSize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertAspectRatio(t *testing.T) {
	tests := []struct {
		name string
		req  openai.ImageRequest
		want string
	}{
		{name: "square default", req: openai.ImageRequest{Size: openai.CreateImageSize1024x1024}, want: "1:1"},
		{name: "landscape", req: openai.ImageRequest{Size: openai.CreateImageSize1792x1024}, want: "16:9"},
		{name: "portrait", req: openai.ImageRequest{Size: openai.CreateImageSize1024x1792}, want: "9:16"},
		{name: "three to two", req: openai.ImageRequest{Size: openai.CreateImageSize1536x1024}, want: "3:2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertAspectRatio(tt.req); got != tt.want {
				t.Fatalf("convertAspectRatio() = %q, want %q", got, tt.want)
			}
		})
	}
}
