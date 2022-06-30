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

package rawparser

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"github.com/buger/jsonparser"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/compressor"
)

func ParseStream(r io.Reader, contentEncoding, cusContentEncoding, format string, callback func(buf []byte) error) error {
	switch contentEncoding {
	case "gzip":
		zr, err := compressor.GetGzipReader(r)
		if err != nil {
			return fmt.Errorf("gzip reader: %w", err)
		}
		defer compressor.PutGzipReader(zr)
		r = zr
	case "snappy":
		zr := compressor.GetSnappyReader(r)
		defer compressor.PutSnappyReader(zr)
		r = zr
	}

	switch cusContentEncoding {
	case "base64":
		r = base64.NewDecoder(base64.StdEncoding, r)
	}

	switch format {
	case "jsonl":
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			buf := scanner.Bytes()
			if lib.IsJSONArray(buf) {
				return fmt.Errorf("is json array not json lines")
			}
			if err := callback(buf); err != nil {
				return fmt.Errorf("callback json line: %w", err)
			}
		}
	default:
		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("io readall: %w", err)
		}

		if !lib.IsJSONArray(data) {
			return fmt.Errorf("invalid data format: not a json array")
		}
		if _, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				log.Printf("rowparser arrayeach err: %s", err)
				return
			}
			if err := callback(value); err != nil {
				log.Printf("rowparser callback err: %s", err)
				return
			}
		}); err != nil {
			return fmt.Errorf("array each: %w", err)
		}
	}
	return nil
}
