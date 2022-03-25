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

package operator

import (
	"regexp"

	"github.com/erda-project/erda/modules/oap/collector/common"
)

type Regex struct {
	cfg     ModifierCfg
	pattern *regexp.Regexp
}

func (r *Regex) Operate(pairs map[string]interface{}) map[string]interface{} {
	val, ok := pairs[r.cfg.Key]
	if !ok {
		return pairs
	}
	for k, v := range common.RegexGroupMap(r.pattern, val.(string)) {
		pairs[k] = v
	}
	return pairs
}

func NewRegex(cfg ModifierCfg) Operator {
	return &Regex{cfg: cfg, pattern: regexp.MustCompile(cfg.Value)}
}
