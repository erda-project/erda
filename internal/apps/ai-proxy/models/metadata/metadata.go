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

package metadata

import (
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	"github.com/erda-project/erda/pkg/strutil"
)

type Metadata struct {
	Public map[string]any `json:"public,omitempty"`
	Secret map[string]any `json:"secret,omitempty"`
}

func (m *Metadata) FromProtobuf(pb *pb.Metadata) {
	*m = Metadata{
		Public: make(map[string]any),
		Secret: make(map[string]any),
	}
	if pb == nil {
		return
	}
	cputil.MustObjJSONTransfer(pb, m)
	if m.Public == nil {
		m.Public = make(map[string]any)
	}
	if m.Secret == nil {
		m.Secret = make(map[string]any)
	}
}

func (m *Metadata) ToJson() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m *Metadata) MergeMap() map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range m.Public {
		result[k] = v
	}
	for k, v := range m.Secret {
		result[k] = v
	}
	return result
}

func (m *Metadata) GetPublicValueByKey(key string, ignoreCaseOpt ...bool) (string, bool) {
	if m == nil {
		return "", false
	}
	var ignoreCase bool
	if len(ignoreCaseOpt) > 0 {
		ignoreCase = ignoreCaseOpt[0]
	}
	for k, v := range m.Public {
		if ignoreCase {
			if strings.EqualFold(key, k) {
				return strutil.String(v), true
			}
		} else {
			if key == k {
				return strutil.String(v), true
			}
		}
	}
	return "", false
}

func (m *Metadata) GetSecretValueByKey(key string, ignoreCaseOpt ...bool) (string, bool) {
	if m == nil {
		return "", false
	}
	var ignoreCase bool
	if len(ignoreCaseOpt) > 0 {
		ignoreCase = ignoreCaseOpt[0]
	}
	for k, v := range m.Secret {
		if ignoreCase {
			if strings.EqualFold(key, k) {
				return strutil.String(v), true
			}
		} else {
			if key == k {
				return strutil.String(v), true
			}
		}
	}
	return "", false
}

func (m *Metadata) GetValueByKey(key string, optionalCfg ...Config) (string, bool) {
	if m == nil {
		return "", false
	}
	cfg := getCfgFromArgs(optionalCfg...)
	if v, ok := m.GetPublicValueByKey(key, cfg.IgnoreCase); ok {
		return v, ok
	}
	return m.GetSecretValueByKey(key, cfg.IgnoreCase)
}

type Config struct {
	IgnoreCase   bool   // default: false
	DefaultValue string // default: ""
}

func getCfgFromArgs(optionalCfg ...Config) Config {
	cfg := Config{
		IgnoreCase:   false,
		DefaultValue: "",
	}
	if len(optionalCfg) > 0 {
		cfg = optionalCfg[0]
	}
	return cfg
}

func (m *Metadata) MustGetValueByKey(key string, optionalCfg ...Config) string {
	v, ok := m.GetValueByKey(key, optionalCfg...)
	if ok {
		return v
	}
	// check default value
	cfg := getCfgFromArgs(optionalCfg...)
	if cfg.DefaultValue != "" {
		return cfg.DefaultValue
	}
	return ""
}

func FromProtobuf(pb *pb.Metadata) Metadata {
	m := new(Metadata)
	m.FromProtobuf(pb)
	return *m
}

func (m *Metadata) ToProtobuf() *pb.Metadata {
	if m == nil {
		return nil
	}
	result := new(pb.Metadata)
	result.Public = make(map[string]*structpb.Value)
	result.Secret = make(map[string]*structpb.Value)
	cputil.MustObjJSONTransfer(m.Public, &result.Public)
	cputil.MustObjJSONTransfer(m.Secret, &result.Secret)
	return result
}
