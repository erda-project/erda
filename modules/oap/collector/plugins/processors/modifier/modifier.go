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

package modifier

type modifierCfg struct {
	Key    string `file:"key"`
	Value  string `file:"value"`
	Action Action `file:"action"`
}

type Action string

const (
	Add    Action = "add"
	Set    Action = "set"
	Drop   Action = "drop"
	Rename Action = "rename"
)

func (p *provider) modify(tags map[string]string) {
	for _, cfg := range p.Cfg.Rules {
		switch cfg.Action {
		case Add:
			if _, ok := tags[cfg.Key]; ok {
				continue
			}
			tags[cfg.Key] = cfg.Value
		case Set:
			tags[cfg.Key] = cfg.Value
		case Drop:
			delete(tags, cfg.Key)
		case Rename:
			if _, ok := tags[cfg.Key]; !ok {
				continue
			}
			tags[cfg.Value] = tags[cfg.Key]
			delete(tags, cfg.Key)
		}
	}
}
