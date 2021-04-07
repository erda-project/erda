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

package apistructs

import (
	"encoding/json"
)

// EventCreateRequest  用于发送 event 的 json request (非 OPENAPI)
// POST: <eventbox>/api/dice/eventbox/message/create
type EventCreateRequest struct {
	EventHeader
	Sender  string
	Content interface{}
}

// MarshalJSON EventCreateRequest 的自定义 marshal 方法
// 将 EventCreateRequest 序列化的结果匹配 internal/eventbox/webhook.EventMessage
// 也就是 使用者只要构造 EventCreateRequest, 而不用去构造 EventMessage
func (r EventCreateRequest) MarshalJSON() ([]byte, error) {
	result := struct {
		Sender  string      `json:"sender"`
		Content interface{} `json:"content"`
		Labels  struct {
			Webhook EventHeader `json:"WEBHOOK"`
		} `json:"labels"`
	}{
		Sender:  r.Sender,
		Content: r.Content,
		Labels: struct {
			Webhook EventHeader `json:"WEBHOOK"`
		}{Webhook: r.EventHeader},
	}
	return json.Marshal(result)
}
