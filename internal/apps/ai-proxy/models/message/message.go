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

package message

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/message/pb"
)

type Message openai.ChatCompletionMessage

func (m *Message) Scan(src any) error {
	if src == nil {
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid src type for message, got %T", src)
	}
	if len(v) == 0 {
		return nil
	}
	return json.Unmarshal(v, m)
}

func (m Message) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func FromProtobuf(pbMsgs []*pb.Message) Messages {
	var result []openai.ChatCompletionMessage
	for _, pbMsg := range pbMsgs {
		result = append(result, openai.ChatCompletionMessage{
			Role:         pbMsg.Role,
			Content:      pbMsg.Content,
			Name:         pbMsg.Name,
			FunctionCall: nil,
		})
	}
	return result
}

type Messages []openai.ChatCompletionMessage

func (msgs *Messages) Scan(src any) error {
	if src == nil {
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid src type for messages, got %T", src)
	}
	if len(v) == 0 {
		return nil
	}
	return json.Unmarshal(v, msgs)
}

func (msgs Messages) Value() (driver.Value, error) {
	return json.Marshal(msgs)
}

func (msgs Messages) ToProtobuf() []*pb.Message {
	var result []*pb.Message
	for _, msg := range msgs {
		result = append(result, &pb.Message{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		})
	}
	return result
}
