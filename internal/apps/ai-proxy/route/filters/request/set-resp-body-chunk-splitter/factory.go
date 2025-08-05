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

package set_resp_body_chunk_splitter

import (
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// isCompatible checks whether the given splitter matches actual Content-Type.
func isCompatible(sp ChunkSplitter, ct string) bool {
	switch sp.(type) {
	case *BedrockStreamSplitter:
		return strings.HasPrefix(ct, "application/vnd.amazon.eventstream")
	case *SSESplitter:
		return strings.HasPrefix(ct, "text/event-stream")
	case *WholeStreamSplitter:
		// WholeStreamSplitter is generic – always compatible
		return true
	case *FixedSizeSplitter:
		return true
	default:
		return true
	}
}

func GetSplitterByResp(resp *http.Response) ChunkSplitter {
	var underlying ChunkSplitter

	// 0) Prefer splitter already stored in context, if compatible
	if resp != nil && resp.Request != nil {
		if sp := ctxhelper.GetRespBodyChunkSplitter(resp.Request.Context()); sp != nil {
			ct := resp.Header.Get("Content-Type")
			if isCompatible(sp, ct) {
				underlying = sp
			}
		}
	}

	// If no suitable splitter is retrieved from context, select based on Content-Type
	if underlying == nil {
		ct := resp.Header.Get("Content-Type")
		switch {
		// Bedrock normal streaming: EventStream
		case strings.HasPrefix(ct, "application/vnd.amazon.eventstream"):
			underlying = &BedrockStreamSplitter{}
		// Regular JSON (including 4xx errors)
		case strings.HasPrefix(ct, "application/json"):
			underlying = &WholeStreamSplitter{}
		// SSE / text types
		case strings.HasPrefix(ct, "text/event-stream"):
			underlying = &SSESplitter{}
		default:
			// Fallback: fixed chunks or whole stream
			underlying = &FixedSizeSplitter{Size: 32 << 10}
		}
	}

	// Check if decompression is needed, first save Content-Encoding to context, then remove response header
	ce := resp.Header.Get("Content-Encoding")
	if ce != "" && ce != "identity" {
		// Save Content-Encoding to context for use during decompression
		ctxhelper.PutResponseContentEncoding(resp.Request.Context(), ce)
		// Remove Content-Encoding header because we will decompress data at framework layer
		resp.Header.Del("Content-Encoding")
	}

	// Put final splitter into ctx
	ctxhelper.PutRespBodyChunkSplitter(resp.Request.Context(), underlying)

	return underlying
}
