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

	"github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
)

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

func (m *Metadata) GetPublicValueByKey(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m.Public[key]
	return v, ok
}

func (m *Metadata) GetSecretValueByKey(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m.Secret[key]
	return v, ok
}

func (m *Metadata) GetValueByKey(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	if v, ok := m.GetPublicValueByKey(key); ok {
		return v, ok
	}
	return m.GetSecretValueByKey(key)
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
