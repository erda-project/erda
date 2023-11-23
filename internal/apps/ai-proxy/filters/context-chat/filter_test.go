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

package context_chat

import (
	"reflect"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/timestamppb"

	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

func Test_getOrderedLimitedChatLogs(t *testing.T) {
	now := time.Now()
	type args struct {
		chatLogs      []*sessionpb.ChatLog
		limitIncluded int
	}
	tests := []struct {
		name string
		args args
		want message.Messages
	}{
		{
			name: "one chat log, not reach limit",
			args: args{
				chatLogs: []*sessionpb.ChatLog{
					{
						RequestAt:  timestamppb.New(now.Add(time.Second)),
						Prompt:     "p1",
						Completion: "c1",
					},
				},
				limitIncluded: 10,
			},
			want: message.Messages{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "p1",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c1",
				},
			},
		},
		{
			name: "one chat log, reach limit 1",
			args: args{
				chatLogs: []*sessionpb.ChatLog{
					{
						RequestAt:  timestamppb.New(now.Add(time.Second)),
						Prompt:     "p1",
						Completion: "c1",
					},
				},
				limitIncluded: 1,
			},
			want: message.Messages{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c1",
				},
			},
		},
		{
			name: "two chat log, not reach limit",
			args: args{
				chatLogs: []*sessionpb.ChatLog{
					{
						RequestAt:  timestamppb.New(now.Add(time.Second * 2)),
						Prompt:     "p2",
						Completion: "c2",
					},
					{
						RequestAt:  timestamppb.New(now.Add(time.Second)),
						Prompt:     "p1",
						Completion: "c1",
					},
				},
				limitIncluded: 4,
			},
			want: message.Messages{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "p1",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c1",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "p2",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c2",
				},
			},
		},
		{
			name: "two chat log, reach limit",
			args: args{
				chatLogs: []*sessionpb.ChatLog{
					{
						RequestAt:  timestamppb.New(now.Add(time.Second * 2)),
						Prompt:     "p2",
						Completion: "c2",
					},
					{
						RequestAt:  timestamppb.New(now.Add(time.Second)),
						Prompt:     "p1",
						Completion: "c1",
					},
				},
				limitIncluded: 3,
			},
			want: message.Messages{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c1",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "p2",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "c2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOrderedLimitedChatLogs(tt.args.chatLogs, tt.args.limitIncluded); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOrderedLimitedChatLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}
