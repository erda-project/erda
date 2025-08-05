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
	"encoding/json"
	"fmt"
	"net/http/httputil"
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
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "context-chat"
)

var (
	_ filter_define.ProxyRequestRewriter = (*SessionContext)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type SessionContext struct {
	SysMsg string `json:"sysMsg" yaml:"sysMsg"`
}

var Creator filter_define.RequestRewriterCreator = func(name string, config json.RawMessage) filter_define.ProxyRequestRewriter {
	var ctx SessionContext
	if err := yaml.Unmarshal(config, &ctx); err != nil {
		panic(fmt.Errorf("failed to unmarshal session context: %v", err))
	}
	return &ctx
}

func (c *SessionContext) OnProxyRequest(pr *httputil.ProxyRequest) error {
	var (
		l  = ctxhelper.MustGetLogger(pr.In.Context())
		db = ctxhelper.MustGetDBClient(pr.In.Context())
	)

	// judge use session-id or prompt-id
	sessionValue, sessionOk := pr.In.Context().Value(ctxhelper.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeySession{})
	promptValue, promptOk := pr.In.Context().Value(ctxhelper.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyPromptTemplate{})

	if !sessionOk && !promptOk {
		return nil
	}
	session, sessionOk := sessionValue.(*sessionpb.Session)
	prompt, promptOk := promptValue.(*promptpb.Prompt)
	if !sessionOk && !promptOk {
		return nil
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
	bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %v", err)
	}
	if err := json.NewDecoder(bodyCopy).Decode(&chatCompletionRequest); err != nil {
		return fmt.Errorf("failed to decode request body, err: %v", err)
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
			return nil
		}
		// get from session's chat-logs
		if session.NumOfCtxMsg > 0 {
			chatLogResp, err := db.SessionClient().GetChatLogs(pr.In.Context(), &sessionpb.SessionChatLogGetRequest{
				SessionId: session.Id,
				PageSize:  uint64(session.NumOfCtxMsg),
				PageNum:   1,
			})
			if err != nil {
				return fmt.Errorf("failed to get session's chat logs, sessionId: %s, err: %v", session.Id, err)
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
	if c.SysMsg != "" {
		systemMessage = &message.Message{Role: openai.ChatMessageRoleSystem, Content: c.SysMsg, Name: "Erda-AI-Assistant"}
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

	// Different models have different bodies, cannot set directly, but put into context for actual model filters to convert
	messageGroup := message.Group{
		AllMessages:             allMessages,
		SystemMessage:           systemMessage,
		SessionTopicMessage:     sessionTopicMessage,
		PromptTemplateMessages:  promptMessages,
		SessionPreviousMessages: sessionPreviousMessages,
		RequestedMessages:       requestedMessages,
	}
	ctxhelper.PutMessageGroup(pr.In.Context(), messageGroup)
	ctxhelper.PutUserPrompt(pr.In.Context(), getPromptFromOpenAIMessage(chatCompletionRequest.Messages[len(chatCompletionRequest.Messages)-1]))
	ctxhelper.PutIsStream(pr.In.Context(), chatCompletionRequest.Stream)

	// set model name in JSON body
	if err := c.trySetJSONBodyModelName(pr); err != nil {
		return fmt.Errorf("failed to set model name in JSON body: %v", err)
	}

	return nil
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

func (c *SessionContext) trySetJSONBodyModelName(pr *httputil.ProxyRequest) error {
	if !strings.HasPrefix(pr.Out.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		return nil
	}
	// update model name
	var reqBody map[string]any
	if err := json.NewDecoder(pr.Out.Body).Decode(&reqBody); err != nil {
		l := ctxhelper.MustGetLogger(pr.Out.Context())
		l.Errorf("failed to decode req body for set json body model name")
		return nil
	}
	model := ctxhelper.MustGetModel(pr.Out.Context())
	var modelName any = model.Name
	if customModelName := model.Metadata.Public["model_name"]; customModelName != nil {
		modelName = customModelName
	}
	reqBody["model"] = modelName
	if err := body_util.SetBody(pr.Out, reqBody); err != nil {
		return fmt.Errorf("failed to set req body for set json body model name, err: %v", err)
	}
	return nil
}
