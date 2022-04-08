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
	"strings"
)

type Add struct {
	cfg ModifierCfg
}

func NewAdd(cfg ModifierCfg) Modifier {
	return &Add{cfg: cfg}
}

func (a *Add) Modify(pairs map[string]interface{}) map[string]interface{} {
	if _, ok := pairs[a.cfg.Key]; ok {
		return pairs
	}
	if a.cfg.Value == "" {
		return pairs
	}
	pairs[a.cfg.Key] = a.cfg.Value
	return pairs
}

type Set struct {
	cfg ModifierCfg
}

func (s *Set) Modify(pairs map[string]interface{}) map[string]interface{} {

	pairs[s.cfg.Key] = s.cfg.Value
	return pairs
}

func NewSet(cfg ModifierCfg) Modifier {
	return &Set{cfg: cfg}
}

type Drop struct {
	cfg ModifierCfg
}

func (d *Drop) Modify(pairs map[string]interface{}) map[string]interface{} {
	delete(pairs, d.cfg.Key)
	return pairs
}

func NewDrop(cfg ModifierCfg) Modifier {
	return &Drop{cfg: cfg}
}

type Rename struct {
	cfg ModifierCfg
}

func (r *Rename) Modify(pairs map[string]interface{}) map[string]interface{} {
	// value is the new key
	if _, ok := pairs[r.cfg.Key]; !ok {
		return pairs
	}
	pairs[r.cfg.Value] = pairs[r.cfg.Key]
	delete(pairs, r.cfg.Key)
	return pairs
}

func NewRename(cfg ModifierCfg) Modifier {
	return &Rename{cfg: cfg}
}

type Copy struct {
	cfg ModifierCfg
}

func (c *Copy) Modify(pairs map[string]interface{}) map[string]interface{} {
	if _, ok := pairs[c.cfg.Key]; !ok {
		return pairs
	}
	pairs[c.cfg.Value] = pairs[c.cfg.Key]
	return pairs
}

func NewCopy(cfg ModifierCfg) Modifier {
	return &Copy{cfg: cfg}
}

type TrimPrefix struct {
	cfg ModifierCfg
}

func (t *TrimPrefix) Modify(pairs map[string]interface{}) map[string]interface{} {
	// key is the prefix
	tmp := make(map[string]interface{}, len(pairs))
	for k, v := range pairs {
		if strings.Index(k, t.cfg.Key) != -1 { // found
			tmp[k[len(t.cfg.Key):]] = v
		} else {
			tmp[k] = v
		}
	}
	return tmp
}

func NewTrimPrefix(cfg ModifierCfg) Modifier {
	return &TrimPrefix{cfg: cfg}
}

type Join struct {
	cfg ModifierCfg
}

func NewJoin(cfg ModifierCfg) Modifier {
	return &Join{cfg: cfg}
}

func (j *Join) Modify(pairs map[string]interface{}) map[string]interface{} {
	if j.cfg.TargetKey == "" {
		return pairs
	}

	vals := make([]string, len(j.cfg.Keys))
	for i, k := range j.cfg.Keys {
		val, ok := pairs[k]
		if !ok {
			return pairs
		}

		valstr, ok := val.(string)
		if !ok {
			return pairs
		}
		vals[i] = valstr
	}
	pairs[j.cfg.TargetKey] = strings.Join(vals, j.cfg.Separator)
	return pairs
}
