// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package form_modal

type Props struct {
	Width      int                    `json:"width,omitempty"`
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Visible    bool                   `json:"visible"`
	Fields     []Field                `json:"fields,omitempty"`
	FormData   map[string]interface{} `json:"formData,omitempty"`
	FormRef    interface{}            `json:"formRef,omitempty"`
	ModalProps map[string]interface{} `json:"modalProps,omitempty"`
}

type Field struct {
	Key            string         `json:"key"`
	Label          string         `json:"label"`
	Required       bool           `json:"required"`
	Component      string         `json:"component"`
	Rules          []FieldRule    `json:"rules,omitempty"`
	ComponentProps ComponentProps `json:"componentProps"`
}

type FieldRule struct {
	Pattern string `json:"pattern"`
	Msg     string `json:"msg"`
}

type ComponentProps struct {
	MaxLength int `json:"maxLength"`
}
