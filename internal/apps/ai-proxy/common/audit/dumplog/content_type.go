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
	"mime"
	"strings"
)

// IsBinaryContentType determines whether content-type is binary stream
func IsBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	// Parse content-type, remove charset and other parameters
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If parsing fails, use the first part of original value (remove content after semicolon)
		if idx := strings.Index(contentType, ";"); idx != -1 {
			mediaType = strings.TrimSpace(contentType[:idx])
		} else {
			mediaType = strings.TrimSpace(contentType)
		}
	}

	mediaType = strings.ToLower(mediaType)

	// Explicit text types - always dump body
	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/x-www-form-urlencoded",
		"application/javascript",
		"application/ecmascript",
		"application/sql",
		"application/graphql",
		"application/ld+json",
		"application/x-ndjson",
	}

	for _, textType := range textTypes {
		if strings.HasPrefix(mediaType, textType) {
			return false
		}
	}

	// Explicit binary types - don't dump body
	binaryTypes := []string{
		"audio/",
		"video/",
		"image/",
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"application/gzip",
		"application/x-tar",
		"application/x-rar-compressed",
		"application/x-7z-compressed",
		"application/msword",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument",
		"application/x-binary",
		"multipart/form-data", // Usually contains binary files
	}

	for _, binaryType := range binaryTypes {
		if strings.HasPrefix(mediaType, binaryType) {
			return true
		}
	}

	// Default case: treat unknown types as text (maintain backward compatibility)
	return false
}

// ShouldDumpBody determines whether to dump body based on content-type
func ShouldDumpBody(contentType string) bool {
	return !IsBinaryContentType(contentType)
}
