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

package common

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/utils"
)

type httpFailureResponse interface {
	StatusCode() int
	ResponseHeader(string) string
}

func formatHTTPFailureFromResponse(action string, resp httpFailureResponse, body []byte) error {
	if resp == nil {
		return formatHTTPFailure(action, 0, "", body, "")
	}
	return formatHTTPFailure(
		action,
		resp.StatusCode(),
		resp.ResponseHeader("Content-Type"),
		body,
		requestIDFromHeaders(resp.ResponseHeader),
	)
}

func formatHTTPFailure(action string, statusCode int, contentType string, body []byte, requestID string) error {
	if parsed, ok := decodeStandardErrorBody(body); ok {
		details := fmt.Sprintf("status=%d, code=%s", statusCode, parsed.Error.Code)
		if requestID != "" {
			details = fmt.Sprintf("%s, request-id=%s", details, requestID)
		}
		return fmt.Errorf("%s", utils.FormatErrMsg(action,
			fmt.Sprintf("%s\ndetails: %s", parsed.Error.Msg, details), false))
	}

	requestIDSuffix := ""
	if requestID != "" {
		requestIDSuffix = fmt.Sprintf(", request-id=%s", requestID)
	}
	return fmt.Errorf("%s", utils.FormatErrMsg(action,
		fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s%s",
			statusCode, contentType, string(body), requestIDSuffix), false))
}

func requestIDFromHeaders(getHeader func(string) string) string {
	candidates := []string{
		"Request-Id",
		"Request-ID",
		"X-Request-Id",
		"X-Request-ID",
		"X-Erda-Request-Id",
		"X-Erda-Request-ID",
		"X-B3-TraceId",
		"X-Trace-Id",
	}
	for _, header := range candidates {
		if value := getHeader(header); value != "" {
			return value
		}
	}
	return ""
}

func decodeStandardErrorBody(body []byte) (*apistructs.Header, bool) {
	var parsed apistructs.Header
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, false
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, false
	}
	if parsed.Error.Code == "" && parsed.Error.Msg == "" {
		return nil, false
	}
	return &parsed, true
}
