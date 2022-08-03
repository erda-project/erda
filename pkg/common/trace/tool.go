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

package trace

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

const (
	compressionFlag    = "0"
	compressionBigText = 1 << 20 // 1MB
)

func BigStringAttribute(key string, data interface{}) attribute.KeyValue {
	text := fmt.Sprintf("%v", data)

	if len(text) < compressionBigText {
		return attribute.String(key, text)
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(text)); err != nil {
		fmt.Println("big string attribute is zip for error:", err)
		return attribute.String(key, text)
	}
	if err := gz.Close(); err != nil {
		fmt.Println("big string attribute is close gzip for error", err)
		return attribute.String(key, text)
	}
	return attribute.String(key, compressionFlag+b.String())
}
