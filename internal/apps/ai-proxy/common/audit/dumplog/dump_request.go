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
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

const headerBodyDelimiter = "\r\n\r\n"

func DumpRequestIn(in *http.Request) {
	logger := ctxhelper.MustGetLoggerBase(in.Context())
	contentType := in.Header.Get("Content-Type")
	shouldDumpBody := ShouldDumpBody(contentType)

	dumpBytesIn, err := httputil.DumpRequest(in, shouldDumpBody)
	if err != nil {
		logger.Warnf("failed to dump request in, err: %v", err)
		return
	}

	dumpString := string(dumpBytesIn)
	if shouldDumpBody && isMultipartFormData(contentType) {
		sanitized, err := sanitizeMultipartDump(dumpBytesIn, contentType)
		if err != nil {
			logger.Warnf("failed to sanitize multipart request body, err: %v", err)
			dumpString = "(multipart body sanitized failed)"
		} else if sanitized != "" {
			dumpString = sanitized
		}
	}

	if shouldDumpBody {
		logger.Debugf("dump proxy request in:\n%s", dumpString)
	} else {
		logger.Debugf("dump proxy request in (body omitted due to binary content-type: %s):\n%s", contentType, dumpString)
	}

	// save to sink
	audithelper.Note(in.Context(), "request_body", dumpString)
}

func DumpRequestOut(out *http.Request) {
	contentType := out.Header.Get("Content-Type")
	shouldDumpBody := ShouldDumpBody(contentType)

	dumpBytesOut, err := httputil.DumpRequestOut(out, shouldDumpBody)
	if err != nil {
		ctxhelper.MustGetLoggerBase(out.Context()).Debugf("failed to dump request out, err: %v", err)
		return
	}

	if shouldDumpBody {
		ctxhelper.MustGetLoggerBase(out.Context()).Debugf("dump proxy request out:\n%s", dumpBytesOut)
	} else {
		ctxhelper.MustGetLoggerBase(out.Context()).Debugf("dump proxy request out (body omitted due to binary content-type: %s):\n%s", contentType, dumpBytesOut)
	}

	// save to sink
	audithelper.Note(out.Context(), "actual_request_body", string(dumpBytesOut))
	audithelper.Note(out.Context(), "actual_request_url", out.URL.String())
}

func sanitizeMultipartDump(dump []byte, contentType string) (string, error) {
	header, body, ok := splitDumpHeaderBody(dump)
	if !ok {
		return "", nil
	}
	formattedBody, err := formatMultipartFormData(contentType, body)
	if err != nil {
		return "", err
	}
	if formattedBody == "" {
		return "", nil
	}

	var builder strings.Builder
	builder.Grow(len(header) + len(formattedBody) + len(headerBodyDelimiter))
	builder.Write(header)
	builder.WriteString(headerBodyDelimiter)
	builder.WriteString(formattedBody)
	return builder.String(), nil
}

func splitDumpHeaderBody(dump []byte) ([]byte, []byte, bool) {
	idx := bytes.Index(dump, []byte(headerBodyDelimiter))
	if idx == -1 {
		return nil, nil, false
	}
	header := dump[:idx]
	body := dump[idx+len(headerBodyDelimiter):]
	return header, body, true
}

func formatMultipartFormData(contentType string, body []byte) (string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(mediaType, "multipart/form-data") {
		return "", nil
	}
	boundary, ok := params["boundary"]
	if !ok {
		return "", fmt.Errorf("missing multipart boundary")
	}
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	var builder strings.Builder
	partCount := 0

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		name := part.FormName()
		if name == "" {
			name = part.FileName()
		}
		if name == "" {
			name = "(unnamed)"
		}

		value := "(binary)"
		partContentType := part.Header.Get("Content-Type")
		if part.FileName() == "" && !IsBinaryContentType(partContentType) {
			data, err := io.ReadAll(part)
			if err != nil {
				return "", err
			}
			value = string(data)
		} else {
			if _, err := io.Copy(io.Discard, part); err != nil {
				return "", err
			}
		}
		if err := part.Close(); err != nil {
			return "", err
		}

		if partCount == 0 {
			builder.WriteString("Form Data:\n")
		}
		builder.WriteString(name)
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteByte('\n')
		partCount++
	}

	if partCount == 0 {
		return "", nil
	}

	return strings.TrimRight(builder.String(), "\n"), nil
}
