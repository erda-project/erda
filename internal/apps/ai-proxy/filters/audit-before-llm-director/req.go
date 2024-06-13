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
	"github.com/erda-project/erda/internal/apps/ai-proxy/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/excerptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *Filter) OnOriginalRequest(ctx context.Context, infor reverseproxy.HttpInfor) {
	var createReq pb.AuditCreateRequestWhenReceived
	createReq.RequestAt = timestamppb.New(time.Now())
	createReq.AuthKey = vars.TrimBearer(infor.Header().Get(vars.TrimBearer(httputil.HeaderKeyAuthorization)))
	// request body
	createReq.RequestBody = excerptor.ExcerptActualRequestBody(infor.BodyBuffer().String())
	// user agent
	createReq.UserAgent = httputil.GetUserAgent(infor.Header())
	// x request id
	createReq.XRequestId = vars.GetFromHeader(infor, vars.XRequestId)

	// metadata
	createReq.RequestContentType = vars.GetFromHeader(infor, httputil.HeaderKeyContentType)
	createReq.RequestHeader = func() string {
		b, err := json.Marshal(infor.Header())
		if err != nil {
			return err.Error()
		}
		return string(b)
	}()

	createReq.IdentityPhoneNumber = vars.GetFromHeader(infor, vars.XAIProxyPhone)
	createReq.IdentityJobNumber = vars.GetFromHeader(infor, vars.XAIProxyJobNumber)

	createReq.DingtalkStaffId = vars.GetFromHeader(infor, vars.XAIProxyDingTalkStaffID)
	createReq.DingtalkChatType = vars.GetFromHeader(infor, vars.XAIProxyChatType)
	createReq.DingtalkChatTitle = vars.GetFromHeader(infor, vars.XAIProxyChatTitle)
	createReq.DingtalkChatId = vars.GetFromHeader(infor, vars.XAIProxyChatId)

	createReq.BizSource = vars.GetFromHeader(infor, vars.XAIProxySource)
	createReq.Username = vars.GetFromHeader(infor, vars.XAIProxyUsername, vars.XAIProxyName)
	createReq.Email = vars.GetFromHeader(infor, vars.XAIProxyEmail)

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

	var updateReq pb.AuditUpdateRequestAfterSpecificContextParsed
	updateReq.AuditId = auditRecID
	// prompt
	prompt, _ := ctxhelper.GetUserPrompt(ctx)
	updateReq.Prompt = prompt

	// metadata, routing by request path
	switch common.GetRequestRoutePath(ctx) {
	case common.RequestPathPrefixV1ChatCompletions, common.RequestPathPrefixV1Completions:
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
	case common.RequestPathPrefixV1Audio:
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
	case common.RequestPathPrefixV1Images:
		imageInfo, ok := ctxhelper.GetImageInfo(ctx)
		if !ok {
			goto Next
		}
		updateReq.ImageQuality = imageInfo.ImageQuality
		updateReq.ImageSize = imageInfo.ImageSize
		updateReq.ImageStyle = imageInfo.ImageStyle
	case common.RequestPathPrefixV1Assistants:

	default:
		// do nothing
	}

Next:

	// update audit into db
	_, err = ctxhelper.MustGetDBClient(ctx).AuditClient().UpdateAfterSpecificContextParsed(ctx, &updateReq)
	if err != nil {
		// log it
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit: %v", err)
	}
	return reverseproxy.Continue, nil
}
