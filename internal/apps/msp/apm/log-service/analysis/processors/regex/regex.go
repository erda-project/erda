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
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/recallsong/go-utils/reflectx"
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
	keys       []*pb.FieldDefine
	appendTags map[string]string
	replaceKey map[string]string
	converts   []func(text string) (interface{}, error)
	pattern    string
	regexps    regexps
}

type regexps struct {
	defaultReg *regexp.Regexp  // default regexp
	zwaReg     *regexp2.Regexp // zero-width assertion regexp
}

// New .
func New(metric string, cfg []byte) (processors.Processor, error) {
	var c config
	err := json.Unmarshal(cfg, &c)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal regexp config: %s", err)
	}
	if len(c.Pattern) == 0 {
		return nil, fmt.Errorf("missing pattern")
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
	p := &processor{
		metric:     metric,
		keys:       c.Keys,
		appendTags: c.AppendTags,
		replaceKey: c.ReplaceKey,
		converts:   converts,
		pattern:    c.Pattern,
	}
	regexps, err := p.initRegexps(c.Pattern)
	if err != nil {
		return nil, err
	}
	p.regexps = regexps
	return p, nil
}

func (p *processor) initRegexps(pattern string) (regexps regexps, err error) {
	// default regexp
	if defaultReg, _err := regexp.Compile(pattern); _err != nil {
		logrus.Warnf("failed to compile regexp pattern(default regexp): %s", _err)
	} else {
		regexps.defaultReg = defaultReg
		return
	}

	// zwa regexp
	if zwaReg, _err := regexp2.Compile(pattern, regexp2.RE2); _err != nil {
		logrus.Warnf("failed to compile regexp pattern(zero-width-assertion ergexp): %s", _err)
	} else {
		regexps.zwaReg = zwaReg
		return
	}

	// not found
	return regexps, fmt.Errorf("no regexp handler available")
}

// ErrNotMatch .
var ErrNotMatch = fmt.Errorf("not match regexp")

// Process .
func (p *processor) Process(content string) (string, map[string]interface{}, map[string]string, map[string]string, error) {
	switch true {
	case p.regexps.zwaReg != nil:
		// handle by zwaReg
		return p.handleByZwaReg(content)
	case p.regexps.defaultReg != nil:
		// handle by defaultReg
		return p.handleByDefaultReg(content)
	default:
		return "", nil, nil, nil, fmt.Errorf("no regexp handler available")
	}
}

func (p *processor) handleByDefaultReg(content string) (string, map[string]interface{}, map[string]string, map[string]string, error) {
	match := p.regexps.defaultReg.FindAllSubmatch(reflectx.StringToBytes(content), 1)
	if len(match) <= 0 {
		return "", nil, nil, nil, ErrNotMatch
	}
	fields := make(map[string]interface{})
	for _, parts := range match {
		if len(parts) != len(p.keys)+1 {
			return "", nil, nil, nil, ErrNotMatch
		}
		for i, byts := range parts[1:] {
			if i < len(p.keys) {
				key := p.keys[i]
				convert := p.converts[i]
				val, err := convert(reflectx.BytesToString(byts))
				if err != nil {
					return "", nil, nil, nil, ErrNotMatch
				}
				fields[key.Key] = val
			}
		}
		break // 只处理第一次匹配
	}
	return p.metric, fields, p.appendTags, p.replaceKey, nil
}

func (p *processor) handleByZwaReg(content string) (string, map[string]interface{}, map[string]string, map[string]string, error) {
	match, err := p.regexps.zwaReg.FindStringMatch(content) // 只处理第一次匹配
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

func (p *processor) Pattern() string {
	return p.pattern
}

func init() {
	processors.RegisterProcessor("regexp", New)
}
