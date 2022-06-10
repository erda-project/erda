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

var _ ObservableData = &Raw{}

// bytes representation of ObservableData for performance
type Raw struct {
	Data []byte            `json:"data"`
	Meta map[string]string `json:"meta"`
}

func (r *Raw) Hash() uint64 {
	return 0
}

func NewRaw(data []byte) *Raw {
	return &Raw{Data: data, Meta: map[string]string{}}
}

func (r *Raw) GetTags() map[string]string {
	return map[string]string{}
}
