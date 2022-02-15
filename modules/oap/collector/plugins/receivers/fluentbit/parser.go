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
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	jsoniter "github.com/json-iterator/go"
)

const defaultLogBatchSize = 10

func (p *provider) convertToLogs(buf []byte) (*model.Logs, error) {
	now := time.Now()
	res := &model.Logs{
		// TODO avoid gc
		Logs: make([]*lpb.Log, 0, defaultLogBatchSize),
	}

	_, err := jsonparser.ArrayEach(buf, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			p.Log.Errorf("json parse err: %v", err)
			return
		}
		lg := &lpb.Log{
			Attributes: make(map[string]string),
		}

		timeStr, err := jsonparser.GetString(value, p.Cfg.FLBKeyMappings.TimestampNano)
		if err != nil {
			p.Log.Infof("get key %s err: %s", p.Cfg.FLBKeyMappings.TimestampNano, err)
		}
		lg.TimeUnixNano = uint64(parseTime(timeStr, now).UnixNano())

		name, _ := jsonparser.GetString(value, p.Cfg.FLBKeyMappings.Name)
		lg.Name = name

		severity, _ := jsonparser.GetString(value, p.Cfg.FLBKeyMappings.Severity)
		lg.Severity = severity

		content, err := jsonparser.GetString(value, p.Cfg.FLBKeyMappings.Content)
		if err != nil {
			p.Log.Errorf("get key %s err: %s", p.Cfg.FLBKeyMappings.Content, err)
			return
		}
		lg.Content = content

		erdaBuf, _, _, err := jsonparser.Get(value, p.Cfg.FLBKeyMappings.Erda)
		if err != nil {
			p.Log.Errorf("get key %s err: %s", p.Cfg.FLBKeyMappings.Erda, err)
			return
		}
		erdaTags := parseMapStr("", erdaBuf)
		for k, v := range erdaTags {
			lg.Attributes[k] = v
		}

		k8sBuf, _, _, _ := jsonparser.Get(value, p.Cfg.FLBKeyMappings.Kubernetes)
		k8sTags := parseMapStr("k8s", k8sBuf)
		for k, v := range k8sTags {
			lg.Attributes[k] = v
		}

		res.Logs = append(res.Logs, lg)
	})
	if err != nil {
		return nil, fmt.Errorf("parser err: %w", err)
	}
	return res, nil
}

func parseTime(data string, now time.Time) time.Time {
	t, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		return now
	}
	return t
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
