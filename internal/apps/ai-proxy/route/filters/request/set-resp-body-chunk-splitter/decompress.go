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
	"io"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// DecompressingChunkSplitter wraps any ChunkSplitter and provides automatic decompression functionality.
// It automatically detects compression format based on HTTP response headers and decompresses,
// ensuring subsequent filters receive decompressed data.
type DecompressingChunkSplitter struct {
	underlying       ChunkSplitter
	resp             *http.Response
	decompressor     io.ReadCloser
	initialized      bool
	decompressNeeded bool
}

func NewDecompressingChunkSplitter(underlying ChunkSplitter, resp *http.Response) *DecompressingChunkSplitter {
	return &DecompressingChunkSplitter{
		underlying: underlying,
		resp:       resp,
	}
}

func (d *DecompressingChunkSplitter) NextChunk(r io.Reader) ([]byte, error) {
	if !d.initialized {
		d.initialized = true

		// Get original Content-Encoding value from context
		ce, ok := ctxhelper.GetResponseContentEncoding(d.resp.Request.Context())
		if !ok {
			ce = d.resp.Header.Get("Content-Encoding")
		}
		if ce != "" && ce != "identity" {
			d.decompressNeeded = true
			// Create a temporary header to pass to decompressor
			tempHeader := make(http.Header)
			tempHeader.Set("Content-Encoding", ce)
			// Create decompressor
			var err error
			d.decompressor, err = NewBodyDecompressor(tempHeader, r)
			if err != nil {
				return nil, err
			}
		} else {
			d.decompressNeeded = false
		}
	}

	// If decompression is needed, use decompressor
	if d.decompressNeeded && d.decompressor != nil {
		return d.underlying.NextChunk(d.decompressor)
	}

	// Otherwise use original data stream directly
	return d.underlying.NextChunk(r)
}
