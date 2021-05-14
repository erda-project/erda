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

package types

import (
	"strings"
)

// LabelKey `Message.Labels' key
type LabelKey string

// Message 通用消息结构体
type Message struct {
	Sender  string                   `json:"sender"`
	Content interface{}              `json:"content"`
	Labels  map[LabelKey]interface{} `json:"labels"`
	Time    int64                    `json:"time,omitempty"` // UnixNano

	originContent interface{} `json:"-"`
}

// Before 是否早于 `t'
func (m *Message) Before(t int64) bool {
	return m.Time < t
}

// OriginContent get `Message.originContent'
func (m *Message) OriginContent() interface{} {
	return m.originContent
}

// SetOriginContent set `Message.originContent'
func (m *Message) SetOriginContent(content interface{}) {
	m.originContent = content
}

// HasPrefix 格式化 labelkey & `s' 之后，判断是否有 `s' 前缀
func (k LabelKey) HasPrefix(s string) bool {
	k_ := k.Normalize()
	s_ := LabelKey(s).Normalize()
	return strings.HasPrefix(string(k_), s_)
}

// Equal 格式化 labelkey & `s' 之后，判断是否相等
func (k LabelKey) Equal(s string) bool {
	k_ := k.Normalize()
	s_ := LabelKey(s).Normalize()
	return s_ == k_
}

// Normalize 格式化 labelkey
func (k LabelKey) Normalize() string {
	if !strings.HasPrefix(string(k), "/") {
		return "/" + string(k)
	}
	return string(k)
}

// NormalizeLabelKey 格式化 labelkey & 转换类型
func (k LabelKey) NormalizeLabelKey() LabelKey {
	return LabelKey(k.Normalize())
}
