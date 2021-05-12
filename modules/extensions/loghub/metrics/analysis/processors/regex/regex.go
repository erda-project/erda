// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package regex

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/recallsong/go-utils/reflectx"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors/convert"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

type config struct {
	Pattern string                 `json:"pattern"`
	Keys    []*metrics.FieldDefine `json:"keys"`
}

type processor struct {
	metric   string
	reg      *regexp.Regexp
	keys     []*metrics.FieldDefine
	converts []func(text string) (interface{}, error)
}

// New .
func New(metric string, cfg []byte) (processors.Processor, error) {
	var c config
	err := json.Unmarshal(cfg, &c)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal regexp config: %s", err)
	}
	reg, err := regexp.Compile(c.Pattern)
	if err != nil {
		return nil, fmt.Errorf("fail to compile regexp pattern: %s", err)
	}
	if len(c.Keys) <= 0 {
		return nil, fmt.Errorf("regexp keys must not be empty")
	}
	converts := make([]func(text string) (interface{}, error), len(c.Keys), len(c.Keys))
	for i, key := range c.Keys {
		if len(key.Key) <= 0 {
			return nil, fmt.Errorf("regexp key must not be empty")
		}
		conv := convert.Converter(key.Type)
		converts[i] = conv
	}
	return &processor{
		metric:   metric,
		reg:      reg,
		keys:     c.Keys,
		converts: converts,
	}, nil
}

// ErrNotMatch .
var ErrNotMatch = fmt.Errorf("not match regexp")

// Process .
func (p *processor) Process(content string) (string, map[string]interface{}, error) {
	match := p.reg.FindAllSubmatch(reflectx.StringToBytes(content), 1)
	if len(match) <= 0 {
		return "", nil, ErrNotMatch
	}
	fields := make(map[string]interface{})
	for _, parts := range match {
		if len(parts) != len(p.keys)+1 {
			return "", nil, ErrNotMatch
		}
		for i, byts := range parts[1:] {
			if i < len(p.keys) {
				key := p.keys[i]
				convert := p.converts[i]
				val, err := convert(reflectx.BytesToString(byts))
				if err != nil {
					return "", nil, ErrNotMatch
				}
				fields[key.Key] = val
			}
		}
		break // 只处理第一次匹配
	}
	return p.metric, fields, nil
}

func (p *processor) Keys() []*metrics.FieldDefine {
	return p.keys
}

func init() {
	processors.RegisterProcessor("regexp", New)
}
