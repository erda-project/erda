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

package resp_body_util

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// toolDescriptionMaxLen is the max chars kept for each tool description in audit.
	toolDescriptionMaxLen = 200
)

// TruncateBodyForAudit truncates body to head+tail bytes with an omission placeholder in between.
func TruncateBodyForAudit(body []byte, headLimit, tailLimit int) []byte {
	if len(body) == 0 {
		return body
	}
	if headLimit < 0 {
		headLimit = 0
	}
	if tailLimit < 0 {
		tailLimit = 0
	}
	if headLimit+tailLimit >= len(body) {
		return body
	}
	if tailLimit > len(body) {
		tailLimit = len(body)
	}
	if headLimit > len(body)-tailLimit {
		headLimit = len(body) - tailLimit
	}
	omittedBytes := len(body) - headLimit - tailLimit
	placeholder := []byte(fmt.Sprintf("...[omitted %d bytes]...", omittedBytes))
	truncated := make([]byte, 0, headLimit+len(placeholder)+tailLimit)
	truncated = append(truncated, body[:headLimit]...)
	truncated = append(truncated, placeholder...)
	if tailLimit > 0 {
		truncated = append(truncated, body[len(body)-tailLimit:]...)
	}
	return truncated
}

// OptimizeBodyForAudit is an SSE-aware audit body reducer.
// For SSE streams it processes events one by one, dropping ephemeral delta events
// and compressing large tool schemas so the stored body remains meaningful.
// For non-SSE bodies it falls back to head+tail byte truncation.
// The body passed in is the raw chunk stream — it does NOT contain HTTP headers.
func OptimizeBodyForAudit(body []byte, headLimit, tailLimit int) []byte {
	if len(body) == 0 {
		return body
	}

	txt := string(body)

	// only apply SSE optimization when there are actual SSE data lines.
	// match either a later line ("\ndata: ") or a stream that starts with "data: ".
	if !strings.Contains(txt, "\ndata: ") && !strings.HasPrefix(txt, "data: ") {
		return TruncateBodyForAudit(body, headLimit, tailLimit)
	}

	// split into individual SSE events (delimited by \n\n)
	rawEvents := strings.Split(txt, "\n\n")
	var kept []string
	for _, raw := range rawEvents {
		raw = strings.TrimRight(raw, "\r\n")
		if raw == "" {
			continue
		}
		optimized, drop := optimizeSSEEvent(raw)
		if !drop {
			kept = append(kept, optimized)
		}
	}

	if len(kept) == 0 {
		// nothing to keep — fall back so we at least have something
		return TruncateBodyForAudit(body, headLimit, tailLimit)
	}

	result := strings.Join(kept, "\n\n") + "\n\n"
	// apply size cap after SSE optimization to guard against large response.done snapshots
	return TruncateBodyForAudit([]byte(result), headLimit, tailLimit)
}

// optimizeSSEEvent decides how to store a single SSE event in the audit body.
// returns (optimizedEvent, drop):
//   - drop=true  → exclude this event from the audit body entirely
//   - drop=false → include optimizedEvent (may be rewritten)
func optimizeSSEEvent(raw string) (string, bool) {
	// extract event: and data: lines
	var eventType, dataLine string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLine = strings.TrimPrefix(line, "data: ")
		}
	}

	// if we have a data: line, try to determine the JSON event type
	jsonType := eventType
	if jsonType == "" && dataLine != "" && dataLine != "[DONE]" {
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(dataLine), &m); err == nil {
			if t, ok := m["type"]; ok {
				_ = json.Unmarshal(t, &jsonType)
			}
		}
	}

	switch jsonType {
	// --- drop: ephemeral incremental events; content is captured in completion field ---
	case "response.text.delta",
		"response.output_text.delta",
		"response.content_part.delta",
		"response.function_call_arguments.delta",
		"response.audio.delta",
		"response.audio_transcript.delta":
		return "", true

	// --- compress: response.created can be enormous due to tool schemas ---
	case "response.created":
		compressed := compressResponseCreatedEvent(raw, dataLine)
		return compressed, false

	// --- keep everything else as-is (response.done, output_item.done, etc.) ---
	default:
		return raw, false
	}
}

// compressResponseCreatedEvent strips verbose tool schemas from the response.created event.
// it keeps tool names and truncated descriptions but drops parameter schemas.
func compressResponseCreatedEvent(raw, dataLine string) string {
	if dataLine == "" {
		return raw
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(dataLine), &obj); err != nil {
		// unparseable (e.g. already truncated) — return as-is
		return raw
	}

	respRaw, ok := obj["response"]
	if !ok {
		return raw
	}
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(respRaw, &resp); err != nil {
		return raw
	}

	toolsRaw, ok := resp["tools"]
	if !ok {
		return raw // no tools, nothing to compress
	}

	// tools is []map for Responses API: [{type, name, description, parameters, ...}]
	var tools []map[string]json.RawMessage
	if err := json.Unmarshal(toolsRaw, &tools); err != nil {
		return raw
	}

	modified := false
	for i, tool := range tools {
		// truncate description
		if descRaw, ok := tool["description"]; ok {
			var desc string
			if err := json.Unmarshal(descRaw, &desc); err == nil && len(desc) > toolDescriptionMaxLen {
				tool["description"], _ = json.Marshal(desc[:toolDescriptionMaxLen] + "...")
				modified = true
			}
		}
		// drop parameters schema entirely — it's the biggest part and not needed for audit review
		if _, ok := tool["parameters"]; ok {
			delete(tool, "parameters")
			modified = true
		}
		// also drop strict/other schema fields
		if _, ok := tool["strict"]; ok {
			delete(tool, "strict")
			modified = true
		}
		tools[i] = tool
	}

	if !modified {
		return raw
	}

	resp["tools"], _ = json.Marshal(tools)
	obj["response"], _ = json.Marshal(resp)
	newDataLine, err := json.Marshal(obj)
	if err != nil {
		return raw
	}

	// reconstruct the event preserving event:/id: lines
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "data: ") {
			lines = append(lines, "data: "+string(newDataLine))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
