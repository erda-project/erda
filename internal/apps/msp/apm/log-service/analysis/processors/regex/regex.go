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

package regex

import (
	"encoding/json"
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/log-service/analysis/processors"
	"github.com/erda-project/erda/internal/apps/msp/apm/log-service/analysis/processors/convert"
)

type config struct {
	Pattern    string            `json:"pattern"`
	Keys       []*pb.FieldDefine `json:"keys"`
	AppendTags map[string]string `json:"appendTags"`
	ReplaceKey map[string]string `json:"replaceKey"`
}

type processor struct {
	metric     string
	reg        *regexp2.Regexp
	keys       []*pb.FieldDefine
	appendTags map[string]string
	replaceKey map[string]string
	converts   []func(text string) (interface{}, error)
}

// New .
func New(metric string, cfg []byte) (processors.Processor, error) {
	var c config
	err := json.Unmarshal(cfg, &c)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal regexp config: %s", err)
	}
	reg, err := regexp2.Compile(c.Pattern, regexp2.RE2)
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
		metric:     metric,
		reg:        reg,
		keys:       c.Keys,
		appendTags: c.AppendTags,
		replaceKey: c.ReplaceKey,
		converts:   converts,
	}, nil
}

// ErrNotMatch .
var ErrNotMatch = fmt.Errorf("not match regexp")

// Process .
func (p *processor) Process(content string) (string, map[string]interface{}, map[string]string, map[string]string, error) {
	match, err := p.reg.FindStringMatch(content) // 只处理第一次匹配
	if err != nil {
		logrus.Errorf("failed to find string match, %v", err)
		return "", nil, nil, nil, ErrNotMatch
	}
	if match == nil || match.GroupCount() <= 0 {
		return "", nil, nil, nil, ErrNotMatch
	}

	fields := make(map[string]interface{})
	for i, group := range match.Groups()[1:] {
		if i < len(p.keys) {
			key := p.keys[i]
			convert := p.converts[i]
			val, err := convert(group.String())
			if err != nil {
				return "", nil, nil, nil, ErrNotMatch
			}
			fields[key.Key] = val
		}
	}
	return p.metric, fields, p.appendTags, p.replaceKey, nil
}

func (p *processor) Keys() []*pb.FieldDefine {
	return p.keys
}

func init() {
	processors.RegisterProcessor("regexp", New)
}
