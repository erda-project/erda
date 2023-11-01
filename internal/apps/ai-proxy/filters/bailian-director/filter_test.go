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

package bailian_director

import (
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

func Test_autoFillQaPair(t *testing.T) {
	type args struct {
		msgs message.Messages
	}
	tests := []struct {
		name string
		args args
		want []*ChatQaMessage
	}{
		{
			name: "no prompt",
			args: args{
				msgs: message.Messages{},
			},
			want: nil,
		},
		{
			name: "one user prompt, not paired",
			args: args{
				msgs: message.Messages{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "hello",
					},
				},
			},
			want: []*ChatQaMessage{
				{
					User: "hello",
					Bot:  botOKMsg.Content,
				},
			},
		},
		{
			name: "one bot prompt, not paired",
			args: args{
				msgs: message.Messages{
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "hello",
					},
				},
			},
			want: []*ChatQaMessage{
				{
					User: "hello",
					Bot:  botOKMsg.Content,
				},
			},
		},
		{
			name: "two user prompt, not paired",
			args: args{
				msgs: message.Messages{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "hello",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "world",
					},
				},
			},
			want: []*ChatQaMessage{
				{
					User: "hello",
					Bot:  botOKMsg.Content,
				},
				{
					User: "world",
					Bot:  botOKMsg.Content,
				},
			},
		},
		{
			name: "two bot prompt, not paired",
			args: args{
				msgs: message.Messages{
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "OK1",
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "OK2",
					},
				},
			},
			want: []*ChatQaMessage{
				{
					User: "OK1",
					Bot:  botOKMsg.Content,
				},
				{
					User: "OK2",
					Bot:  botOKMsg.Content,
				},
			},
		},
		{
			name: "bot & user prompt, not paired",
			args: args{
				msgs: message.Messages{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "you are qwenv1",
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "OK1",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "hello",
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "OK2",
					},
				},
			},
			want: []*ChatQaMessage{
				{
					User: "you are qwenv1",
					Bot:  botOKMsg.Content,
				},
				{
					User: "OK1",
					Bot:  botOKMsg.Content,
				},
				{
					User: "hello",
					Bot:  "OK2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := autoFillQaPair(tt.args.msgs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("autoFillQaPair() = %v, want %v", got, tt.want)
			}
		})
	}
}
