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

package form

type Props struct {
	Width      int                    `json:"width"`
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Visible    bool                   `json:"visible"`
	Fields     []interface{}          `json:"fields"`
	FormData   map[string]interface{} `json:"formData"`
	FormRef    interface{}            `json:"formRef"`
	ModalProps map[string]interface{} `json:"modalProps"`
}
