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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"

	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "bailian-director"
)

var (
	_ reverseproxy.RequestFilter = (*BailianDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type BailianDirector struct {
	*reverseproxy.DefaultResponseFilter

	lastCompletionDataLineIndex int
	lastCompletionDataLineText  string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	return &BailianDirector{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
	}, nil
}

func (f *BailianDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *BailianDirector) Enable(ctx context.Context, req *http.Request) bool {
	prov, ok := ctxhelper.GetModelProvider(ctx)
	return ok && prov.Type == modelproviderpb.ModelProviderType_AliyunBailian.String()
}

func (f *BailianDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	prov, _ := ctxhelper.GetModelProvider(ctx)
	model, _ := ctxhelper.GetModel(ctx)
	modelMeta := getModelMeta(model)
	messageGroup, _ := ctxhelper.GetMessageGroup(ctx)

	// use go sdk
	client := fetchClient(prov)
	token, err := client.GetToken()
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to get token, err: %v", err)
	}
	reverseproxy.AppendDirectors(ctx, func(r *http.Request) {
		// rewrite url
		bailianURL := fmt.Sprintf("%s/v2/app/completions", BroadscopeBailianEndpoint)
		u, _ := url.Parse(bailianURL)
		r.URL = u
		r.Host = u.Host
		// rewrite authorization header
		r.Header.Set(httputil.HeaderKeyContentType, string(httputil.ApplicationJsonUTF8))
		r.Header.Set(httputil.HeaderKeyAuthorization, vars.ConcatBearer(token))
		r.Header.Set(httputil.HeaderKeyAccept, string(httputil.ApplicationJsonUTF8))
		r.Header.Del(httputil.HeaderKeyAcceptEncoding) // remove gzip. Actual test: gzip is not ok; deflate is ok; br is ok
	})

	// parse original request body
	var openaiReq openai.ChatCompletionRequest
	if err := json.NewDecoder(infor.BodyBuffer()).Decode(&openaiReq); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to parse request body as openai format, err: %v", err)
	}

	var prompt string
	var historyMsgs []*ChatQaMessage

	if len(openaiReq.Messages) == 0 {
		return reverseproxy.Intercept, fmt.Errorf("no prompt provided")
	}
	lastMsgIndex := len(openaiReq.Messages) - 1
	prompt = openaiReq.Messages[lastMsgIndex].Content
	if messageGroup != nil {
		historyMsgs = transferHistoryMessages(*messageGroup)
	}

	bailianReq := CompletionRequest{
		AppId:   &modelMeta.Secret.AppId,
		Prompt:  &prompt,
		History: historyMsgs,
		Stream:  openaiReq.Stream,
	}
	b, err := json.Marshal(&bailianReq)
	if err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	infor.SetBody(io.NopCloser(bytes.NewBuffer(b)), int64(len(b)))

	return reverseproxy.Continue, nil
}

func transferHistoryMessages(g message.Group) []*ChatQaMessage {
	var qas []*ChatQaMessage

	// system
	if g.SystemMessage != nil {
		qas = append(qas, &ChatQaMessage{
			User: "background",
			Bot:  g.SystemMessage.Content,
		})
	}

	// session topic
	if g.SessionTopicMessage != nil {
		qas = append(qas, &ChatQaMessage{
			User: "session topic",
			Bot:  g.SessionTopicMessage.Content,
		})
	}

	// prompt template
	for _, msg := range g.PromptTemplateMessages {
		qas = append(qas, &ChatQaMessage{
			User: msg.Content,
			Bot:  botOKMsg.Content,
		})
	}

	// session previous
	qas = append(qas, autoFillQaPair(g.SessionPreviousMessages)...)

	// requested, if there are more than one message, only the last one is prompt, others are history
	if len(g.RequestedMessages) > 1 {
		qas = append(qas, autoFillQaPair(g.RequestedMessages[:len(g.RequestedMessages)-1])...)
	}

	return qas
}

var botOKMsg = openai.ChatCompletionMessage{
	Role:    openai.ChatMessageRoleAssistant,
	Content: "Got it",
}

func autoFillQaPair(msgs message.Messages) []*ChatQaMessage {
	if len(msgs) == 0 {
		return nil
	}

	var filledMsgs message.Messages
	for i := 0; i < len(msgs); i++ {
		j := i + 1
		if j >= len(msgs) { // out of range
			filledMsgs = append(filledMsgs, msgs[i], botOKMsg)
			break
		}
		currentMsg, nextMsg := msgs[i], msgs[j]
		if currentMsg.Role == openai.ChatMessageRoleUser && nextMsg.Role == openai.ChatMessageRoleAssistant {
			filledMsgs = append(filledMsgs, currentMsg, nextMsg)
			i = j
			continue
		}
		// not user, just add bot ok msg to pair it
		filledMsgs = append(filledMsgs, currentMsg, botOKMsg)
	}

	// transfer to qa pair
	var result []*ChatQaMessage
	for i := 0; i < len(filledMsgs); i += 2 {
		result = append(result, &ChatQaMessage{
			User: filledMsgs[i].Content,
			Bot:  filledMsgs[i+1].Content,
		})
	}
	return result
}
