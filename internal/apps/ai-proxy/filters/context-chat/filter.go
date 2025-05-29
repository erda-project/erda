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
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"gopkg.in/yaml.v3"

	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "context-chat"
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
	promptValue, promptOk := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyPromptTemplate{})

	if !sessionOk && !promptOk {
		return reverseproxy.Continue, nil
	}
	session, sessionOk := sessionValue.(*sessionpb.Session)
	prompt, promptOk := promptValue.(*promptpb.Prompt)
	if !sessionOk && !promptOk {
		return reverseproxy.Continue, nil
	}

	var allMessages message.Messages
	var systemMessage *message.Message
	var sessionTopicMessage *message.Message
	var promptMessages message.Messages
	var sessionPreviousMessages message.Messages
	var requestedMessages message.Messages

	var chatCompletionRequest openai.ChatCompletionRequest

	// init `JSONSchema.Schema` for `json.Decode`, otherwise, it will report an error
	chatCompletionRequest.ResponseFormat = &openai.ChatCompletionResponseFormat{
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Schema: &jsonschema.Definition{},
		},
	}
	if err := json.NewDecoder(infor.BodyBuffer()).Decode(&chatCompletionRequest); err != nil {
		l.Errorf("failed to decode request body, err: %v", err)
		return reverseproxy.Intercept, err
	}
	if chatCompletionRequest.ResponseFormat.Type == "" {
		chatCompletionRequest.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeText
	}
	if chatCompletionRequest.ResponseFormat.Type != openai.ChatCompletionResponseFormatTypeJSONSchema {
		chatCompletionRequest.ResponseFormat.JSONSchema = nil
	}
	for _, msg := range chatCompletionRequest.Messages {
		// handle user message, wrap by '|start| your question here |end|'
		// to avoid from content-filter
		//if msg.Role == openai.ChatMessageRoleUser {
		//	if msg.Content != "" {
		//		msg.Content = vars.WrapUserPrompt(msg.Content)
		//	} else {
		//		for i, part := range msg.MultiContent {
		//			if part.Text != "" {
		//				msg.MultiContent[i].Text = vars.WrapUserPrompt(part.Text)
		//			}
		//		}
		//	}
		//}
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
			sessionPreviousMessages = getOrderedLimitedChatLogs(chatLogResp.List, int(session.NumOfCtxMsg))

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
		systemMessage = &message.Message{Role: openai.ChatMessageRoleSystem, Content: c.Config.SysMsg, Name: "Erda-AI-Assistant"}
		allMessages = append(allMessages, *systemMessage.ToOpenAI())
	}

	// 1. add session topic
	if sessionTopicMessage != nil {
		allMessages = append(allMessages, *sessionTopicMessage.ToOpenAI())
	}
	// 2. add prompt messages
	allMessages = append(allMessages, promptMessages...)
	// 3. add session chat-logs
	allMessages = append(allMessages, sessionPreviousMessages...)
	// 4. add requested messages
	allMessages = append(allMessages, requestedMessages...)

	// 不同的模型，body 不同，不能直接 set，而是塞入上下文，由真正的 model filters 进行转换
	messageGroup := message.Group{
		AllMessages:             allMessages,
		SystemMessage:           systemMessage,
		SessionTopicMessage:     sessionTopicMessage,
		PromptTemplateMessages:  promptMessages,
		SessionPreviousMessages: sessionPreviousMessages,
		RequestedMessages:       requestedMessages,
	}
	ctxhelper.PutMessageGroup(ctx, messageGroup)
	ctxhelper.PutUserPrompt(ctx, getPromptFromOpenAIMessage(chatCompletionRequest.Messages[len(chatCompletionRequest.Messages)-1]))
	ctxhelper.PutIsStream(ctx, chatCompletionRequest.Stream)

	// update model name
	var reqBody map[string]any
	json.NewDecoder(infor.Body()).Decode(&reqBody)
	reqBody["model"] = ctxhelper.MustGetModel(ctx).Name
	b, _ := json.Marshal(&reqBody)
	infor.SetBody2(b)

	return reverseproxy.Continue, nil
}

func getOrderedLimitedChatLogs(chatLogs []*sessionpb.ChatLog, limitIncluded int) message.Messages {
	// sort by requestAt, oldest first
	sort.Slice(chatLogs, func(i, j int) bool {
		return chatLogs[i].RequestAt.AsTime().Before(chatLogs[j].RequestAt.AsTime())
	})
	// convert to messages
	var limitedChatLogMessages message.Messages
	for _, chatLog := range chatLogs {
		limitedChatLogMessages = append(limitedChatLogMessages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: chatLog.Prompt,
			},
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: chatLog.Completion,
			},
		)
	}
	// cut down to session.NumOfCtxMsg
	if len(limitedChatLogMessages) > limitIncluded {
		limitedChatLogMessages = limitedChatLogMessages[len(limitedChatLogMessages)-limitIncluded:]
	}
	return limitedChatLogMessages
}

func getPromptFromOpenAIMessage(msg openai.ChatCompletionMessage) string {
	if len(msg.MultiContent) > 0 {
		// combine multi-content, only record text information
		var multiTexts []string
		for _, content := range msg.MultiContent {
			if content.Type == openai.ChatMessagePartTypeText && content.Text != "" {
				multiTexts = append(multiTexts, vars.UnwrapUserPrompt(content.Text))
			}
		}
		return strings.Join(multiTexts, "\n")
	}
	return vars.UnwrapUserPrompt(msg.Content)
}
