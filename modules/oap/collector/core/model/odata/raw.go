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

package odata

import (
	"fmt"
)

// bytes representation of ObservableData for performance
type Raws []*Raw

type Raw struct {
	Item []byte    `json:"item"`
	Meta *Metadata `json:"meta"`
}

func NewRaw(item []byte) *Raw {
	return &Raw{Item: item, Meta: &Metadata{Data: map[string]string{}}}
}

func (r *Raw) AddMetadata(key, value string) {
	r.Meta.Add(key, value)
}

func (r *Raw) GetMetadata(key string) (string, bool) {
	return r.Meta.Get(key)
}

func (r *Raw) HandleAttributes(_ func(attr map[string]string) map[string]string) {}

func (r *Raw) HandleName(_ func(name string) string) {}

func (r *Raw) Clone() ObservableData {
	item := make([]byte, len(r.Item))
	copy(item, r.Item)
	return &Raw{
		Item: item,
		Meta: r.Meta.Clone(),
	}
}

func (r *Raw) Source() interface{} {
	return r.Item
}

func (r *Raw) SourceCompatibility() interface{} {
	return r.Item
}

func (r *Raw) SourceType() SourceType {
	return RawType
}

func (r *Raw) String() string {
	return fmt.Sprintf("raw data size => %d", len(r.Item))
}
