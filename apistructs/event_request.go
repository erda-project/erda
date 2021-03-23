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
