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

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	"strings"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
)

type (
	ClientMeta struct {
		Public ClientMetaPublic `json:"public,omitempty"`
		Secret ClientMetaSecret `json:"secret,omitempty"`
	}
	ClientMetaPublic struct {
		DefaultModelIdOfTextGeneration string `json:"default_model_id_of_text_generation,omitempty"`
		DefaultModelIdOfImage          string `json:"default_model_id_of_image,omitempty"`
		DefaultModelIdOfAudio          string `json:"default_model_id_of_audio,omitempty"`
		DefaultModelIdOfEmbedding      string `json:"default_model_id_of_embedding,omitempty"`
		DefaultModelIdOfTextModeration string `json:"default_model_id_of_text_moderation,omitempty"`
	}
	ClientMetaSecret struct {
	}
)

func (m *Metadata) ToClientMeta() (*ClientMeta, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Metadata to json: %v", err)
	}
	var result ClientMeta
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string to ClientMeta: %v", err)
	}
	return &result, nil
}

func (p *ClientMetaPublic) GetDefaultModelIdByModelType(modelType modelpb.ModelType) (string, bool) {
	if p == nil {
		return "", false
	}
	b, _ := json.Marshal(p)
	m := make(map[string]string)
	_ = json.Unmarshal(b, &m)
	defaultModelId, ok := m["default_model_id_of_"+modelType.String()]
	return defaultModelId, ok
}

type (
	ModelProviderMeta struct {
		Public ModelProviderMetaPublic `json:"public,omitempty"`
		Secret ModelProviderMetaSecret `json:"secret,omitempty"`
	}
	ModelProviderMetaPublic struct {
		Scheme   string `json:"scheme,omitempty"`
		Host     string `json:"host,omitempty"`
		Location string `json:"location,omitempty"`
		Region   string `json:"region,omitempty"`
	}
	ModelProviderMetaSecret struct {
		APIKey string `json:"apiKey,omitempty"`
	}
)

func (m *Metadata) ToModelProviderMeta() (*ModelProviderMeta, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Metadata to json: %v", err)
	}
	var result ModelProviderMeta
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal string to ModelProviderMeta: %v", err)
	}
	return &result, nil
}

type Metadata struct {
	Public map[string]string `json:"public,omitempty"`
	Secret map[string]string `json:"secret,omitempty"`
}

func (m *Metadata) FromProtobuf(pb *pb.Metadata) {
	*m = Metadata{
		Public: make(map[string]string),
		Secret: make(map[string]string),
	}
	if pb == nil {
		return
	}
	// public
	for k, v := range pb.Public {
		m.Public[k] = v
	}
	// secret
	for k, v := range pb.Secret {
		m.Secret[k] = v
	}
}

func (m *Metadata) ToJson() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m *Metadata) MergeMap() map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string)
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
				return v, true
			}
		} else {
			if key == k {
				return v, true
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
				return v, true
			}
		} else {
			if key == k {
				return v, true
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

func MetadataFromProtobuf(pb *pb.Metadata) Metadata {
	m := new(Metadata)
	m.FromProtobuf(pb)
	return *m
}

func (m *Metadata) ToProtobuf() *pb.Metadata {
	if m == nil {
		return nil
	}
	result := new(pb.Metadata)
	result.Public = make(map[string]string)
	result.Secret = make(map[string]string)
	for k, v := range m.Public {
		result.Public[k] = v
	}
	for k, v := range m.Secret {
		result.Secret[k] = v
	}
	return result
}

func (m *Metadata) Scan(src any) error {
	if src == nil {
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid src type for metadata, got %T", src)
	}
	if len(v) == 0 {
		return nil
	}
	return json.Unmarshal(v, m)
}

func (m Metadata) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
