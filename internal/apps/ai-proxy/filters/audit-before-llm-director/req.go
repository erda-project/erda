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

package audit_before_llm_director

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *Filter) OnOriginalRequest(ctx context.Context, infor reverseproxy.HttpInfor) {
	var createReq pb.AuditCreateRequestWhenReceived
	createReq.RequestAt = timestamppb.New(time.Now())
	createReq.AuthKey = vars.TrimBearer(infor.Header().Get(vars.TrimBearer(httputil.HeaderKeyAuthorization)))
	// request body
	createReq.RequestBody = infor.BodyBuffer().String()
	// user agent
	createReq.UserAgent = httputil.GetUserAgent(infor.Header())
	// x request id
	createReq.XRequestId = getFromHeader(infor, vars.XRequestId)

	// metadata
	createReq.RequestContentType = getFromHeader(infor, httputil.HeaderKeyContentType)
	createReq.RequestHeader = func() string {
		b, err := json.Marshal(infor.Header())
		if err != nil {
			return err.Error()
		}
		return string(b)
	}()

	createReq.IdentityPhoneNumber = getFromHeader(infor, vars.XAIProxyPhone)
	createReq.IdentityJobNumber = getFromHeader(infor, vars.XAIProxyJobNumber)

	createReq.DingtalkStaffId = getFromHeader(infor, vars.XAIProxyDingTalkStaffID)
	createReq.DingtalkChatType = getFromHeader(infor, vars.XAIProxyChatType)
	createReq.DingtalkChatTitle = getFromHeader(infor, vars.XAIProxyChatTitle)
	createReq.DingtalkChatId = getFromHeader(infor, vars.XAIProxyChatId)

	createReq.BizSource = getFromHeader(infor, vars.XAIProxySource)
	createReq.Username = getFromHeader(infor, vars.XAIProxyUsername, vars.XAIProxyName)
	createReq.Email = getFromHeader(infor, vars.XAIProxyEmail)

	// insert audit into db
	newAudit, err := ctxhelper.MustGetDBClient(ctx).AuditClient().CreateWhenReceived(ctx, &createReq)
	if err != nil {
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to create audit: %v", err)
	}
	if newAudit != nil {
		ctxhelper.PutAuditID(ctx, newAudit.Id)
	}
	return
}

func (f *Filter) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	return f.OnRequestBeforeLLMDirector(ctx, w, infor)
}

func (f *Filter) OnRequestBeforeLLMDirector(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	auditRecID, ok := ctxhelper.GetAuditID(ctx)
	if !ok || auditRecID == "" {
		return
	}

	var updateReq pb.AuditUpdateRequestAfterContextParsed
	updateReq.AuditId = auditRecID
	// prompt
	prompt, _ := ctxhelper.GetUserPrompt(ctx)
	updateReq.Prompt = prompt

	// client id
	client, _ := ctxhelper.GetClient(ctx)
	updateReq.ClientId = client.Id
	// model id
	model, _ := ctxhelper.GetModel(ctx)
	updateReq.ModelId = model.Id
	// session id
	session, _ := ctxhelper.GetSession(ctx)
	if session != nil {
		updateReq.SessionId = session.Id
	}

	// biz source
	updateReq.BizSource = getFromHeader(infor, vars.XAIProxySource)
	// operation id
	updateReq.OperationId = infor.Method() + " " + infor.URL().Path

	// metadata, routing by model type
	switch model.Type {
	case modelpb.ModelType_text_generation:
		var openaiReq openai.ChatCompletionRequest
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&openaiReq); err != nil {
			goto Next
		}
		updateReq.RequestFunctionCallName = func() string {
			// switch type
			switch openaiReq.FunctionCall.(type) {
			case string:
				return openaiReq.FunctionCall.(string)
			case map[string]interface{}:
				var reqFuncCall openai.FunctionCall
				cputil.MustObjJSONTransfer(openaiReq.FunctionCall, &reqFuncCall)
				return reqFuncCall.Name
			case openai.FunctionCall:
				return openaiReq.FunctionCall.(openai.FunctionCall).Name
			case nil:
				return ""
			default:
				return fmt.Sprintf("%v", openaiReq.FunctionCall)
			}
		}()
	case modelpb.ModelType_audio:
		audioInfo, ok := ctxhelper.GetAudioInfo(ctx)
		if !ok {
			goto Next
		}
		updateReq.AudioFileName = audioInfo.FileName
		updateReq.AudioFileSize = audioInfo.FileSize.String()
		updateReq.AudioFileHeaders = func() string {
			b, err := json.Marshal(audioInfo.FileHeaders)
			if err != nil {
				return err.Error()
			}
			return string(b)
		}()
	default:
		// do nothing
	}

Next:

	// set from client token
	setUserInfoFromClientToken(ctx, infor, &updateReq)

	// update audit into db
	_, err = ctxhelper.MustGetDBClient(ctx).AuditClient().UpdateAfterContextParsed(ctx, &updateReq)
	if err != nil {
		// log it
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit: %v", err)
	}
	return reverseproxy.Continue, nil
}

func getFromHeader(infor reverseproxy.HttpInfor, keys ...string) string {
	for _, key := range keys {
		if v := vars.TryUnwrapBase64(infor.Header().Get(key)); v != "" {
			return v
		}
	}
	return ""
}

func setUserInfoFromClientToken(ctx context.Context, infor reverseproxy.HttpInfor, updateReq *pb.AuditUpdateRequestAfterContextParsed) {
	clientToken, ok := ctxhelper.GetClientToken(ctx)
	if !ok || clientToken == nil {
		return
	}
	meta := metadata.FromProtobuf(clientToken.Metadata)
	metaCfg := metadata.Config{IgnoreCase: true}
	updateReq.DingtalkStaffId = meta.MustGetValueByKey(vars.XAIProxyDingTalkStaffID, metaCfg)
	updateReq.Email = meta.MustGetValueByKey(vars.XAIProxyEmail, metaCfg)
	updateReq.IdentityJobNumber = meta.MustGetValueByKey(vars.XAIProxyJobNumber, metaCfg)
	updateReq.Username = meta.MustGetValueByKey(vars.XAIProxyName, metaCfg)
	updateReq.IdentityPhoneNumber = meta.MustGetValueByKey(vars.XAIProxyPhone, metaCfg)
	if getFromHeader(infor, vars.XAIProxySource) == "" { // use token's client's name
		client, ok := ctxhelper.GetClient(ctx)
		if ok {
			updateReq.BizSource = client.Name
		}
	}
}
