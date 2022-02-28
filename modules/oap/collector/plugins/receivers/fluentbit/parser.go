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

package fluentbit

import (
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
	jsoniter "github.com/json-iterator/go"
)

func parseItem(value []byte, cfg flbKeyMappings) (*lpb.Log, error) {
	lg := &lpb.Log{
		Attributes: make(map[string]string),
	}

	timeStr, err := jsonparser.GetString(value, cfg.TimeUnixNano)
	if err != nil {
		return nil, fmt.Errorf("get timeStr from %s, err: %w", cfg.TimeUnixNano, err)
	}
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return nil, fmt.Errorf("parse timeStr err: %w", err)
	}
	lg.TimeUnixNano = uint64(t.UnixNano())

	name, _ := jsonparser.GetString(value, cfg.Name)
	lg.Name = name

	severity, _ := jsonparser.GetString(value, cfg.Severity)
	lg.Severity = severity

	content, err := jsonparser.GetString(value, cfg.Content)
	if err != nil {
		return nil, fmt.Errorf("get key %s err: %w", cfg.Content, err)
	}
	lg.Content = content

	stream, err := jsonparser.GetString(value, cfg.Stream)
	if err != nil {
		stream = "stdout"
	}
	lg.Attributes["stream"] = stream

	k8sBuf, _, _, _ := jsonparser.Get(value, cfg.Kubernetes)
	k8sTags := parseMapStr(cfg.Kubernetes, k8sBuf)
	for k, v := range k8sTags {
		lg.Attributes[k] = v
	}
	return lg, nil
}

func parseMapStr(prefix string, data []byte) map[string]string {
	src := make(map[string]interface{})
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, &src)
	if err != nil {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	flattenMapStr(prefix, src, dst)
	return dst
}

func flattenMapStr(prefix string, src map[string]interface{}, dst map[string]string) {
	if len(prefix) > 0 {
		prefix += "_"
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			flattenMapStr(prefix+k, child, dst)
		case string:
			dst[prefix+k] = v.(string)
		}
	}
}
