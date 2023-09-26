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

package message_context

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "message-context"
)

var (
	_ reverseproxy.RequestFilter = (*SessionContext)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type SessionContext struct {
	Config *Config
}

type Config struct {
	SysMsg string `json:"sysMsg" yaml:"sysMsg"`
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &SessionContext{Config: &cfg}, nil
}

func (c *SessionContext) Enable(_ context.Context, req *http.Request) bool {
	return true
}

func (c *SessionContext) OnRequest(ctx context.Context, _ http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var (
		l  = ctxhelper.GetLogger(ctx)
		db = ctx.Value(vars.CtxKeyDAO{}).(dao.DAO)
	)

	// judge use session-id or prompt-id
	sessionValue, sessionOk := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeySession{})
	promptValue, promptOk := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyPrompt{})

	if !sessionOk && !promptOk {
		return reverseproxy.Continue, nil
	}
	session, sessionOk := sessionValue.(*sessionpb.Session)
	prompt, promptOk := promptValue.(*promptpb.Prompt)
	if !sessionOk && !promptOk {
		return reverseproxy.Continue, nil
	}

	var allMessages message.Messages
	var sessionTopicMessage *message.Message
	var promptMessages message.Messages
	var sessionPreviousMessages message.Messages
	var requestedMessages message.Messages

	var chatCompletionRequest openai.ChatCompletionRequest
	if err := json.NewDecoder(infor.BodyBuffer()).Decode(&chatCompletionRequest); err != nil {
		l.Errorf("failed to decode request body, err: %v", err)
		return reverseproxy.Intercept, err
	}
	for _, msg := range chatCompletionRequest.Messages {
		// handle user message, wrap by '|start| your question here |end|'
		// to avoid from content-filter
		if msg.Role == openai.ChatMessageRoleUser {
			msg.Content = vars.WrapUserPrompt(msg.Content)
		}
		requestedMessages = append(requestedMessages, msg)
	}

	if session != nil {
		if session.IsArchived {
			l.Infof("session %s archived", session.Id)
			return reverseproxy.Continue, nil
		}
		// get from session's chat-logs
		if session.NumOfCtxMsg > 0 {
			chatLogResp, err := db.SessionClient().GetChatLogs(ctx, &sessionpb.SessionChatLogGetRequest{
				SessionId: session.Id,
				PageSize:  uint64(session.NumOfCtxMsg),
				PageNum:   1,
			})
			if err != nil {
				l.Errorf("failed to get session's chat logs, sessionId: %s, err: %v", session.Id, err)
				return reverseproxy.Intercept, err
			}
			// reverse the chat logs
			sort.Slice(chatLogResp.List, func(i, j int) bool {
				return chatLogResp.List[i].RequestAt.AsTime().Before(chatLogResp.List[j].RequestAt.AsTime())
			})
			var num int64
			for _, chatLog := range chatLogResp.List {
				if num > session.NumOfCtxMsg {
					break
				}
				sessionPreviousMessages = append(sessionPreviousMessages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: chatLog.Completion,
				})
				if num > session.NumOfCtxMsg {
					break
				}
				sessionPreviousMessages = append(sessionPreviousMessages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: chatLog.Prompt,
				})
			}
			// reverse previousMessages to make it in order
			strutil.ReverseSlice(sessionPreviousMessages)

			// add topic
			if session.Topic != "" {
				sessionTopicMessage = &message.Message{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Topic: " + session.Topic,
				}
			}
		}
	}

	if prompt != nil {
		promptMessages = message.FromProtobuf(prompt.Messages)
	}

	// compose all messages

	// 0. add system message
	if c.Config.SysMsg != "" {
		allMessages = append(allMessages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: c.Config.SysMsg, Name: "Erda-AI-Assistant"})
	}

	// 1. add session topic
	if sessionTopicMessage != nil {
		allMessages = append(allMessages, openai.ChatCompletionMessage(*sessionTopicMessage))
	}
	// 2. add prompt messages
	allMessages = append(allMessages, promptMessages...)
	// 3. add session chat-logs
	allMessages = append(allMessages, sessionPreviousMessages...)
	// 4. add requested messages
	allMessages = append(allMessages, requestedMessages...)

	// set to request body
	chatCompletionRequest.Messages = allMessages
	b, err := json.Marshal(&chatCompletionRequest)
	if err != nil {
		l.Errorf("failed to marshal request body, err: %v", err)
		return reverseproxy.Intercept, err
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(b)), int64(len(b)))

	return reverseproxy.Continue, nil
}
