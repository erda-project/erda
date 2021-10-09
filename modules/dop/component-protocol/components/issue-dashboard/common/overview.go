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

package common

type Overview struct {
	Props OverviewProps `json:"props,omitempty"`
}

type OverviewProps struct {
	RenderType string        `json:"renderType,omitempty"`
	Value      OverviewValue `json:"value,omitempty"`
}

type OverviewValue struct {
	Direction string         `json:"direction,omitempty"`
	Text      []OverviewText `json:"text,omitempty"`
}

type OverviewText struct {
	Text        string      `json:"text,omitempty"`
	StyleConfig StyleConfig `json:"styleConfig,omitempty"`
}

type StyleConfig struct {
	FontSize int    `json:"fontSize,omitempty"`
	Bold     bool   `json:"bold,omitempty"`
	Color    string `json:"color,omitempty"`
}
