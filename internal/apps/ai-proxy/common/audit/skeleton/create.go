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

package skeleton

import (
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

func CreateSkeleton(in *http.Request) error {
	var createReq pb.AuditCreateRequestWhenReceived
	createReq.RequestAt = timestamppb.New(time.Now())
	createReq.AuthKey = vars.TrimBearer(in.Header.Get(vars.TrimBearer(httperrorutil.HeaderKeyAuthorization)))

	// user agent
	createReq.UserAgent = httperrorutil.GetUserAgent(in.Header)
	// x-request-id
	createReq.XRequestId = ctxhelper.MustGetRequestID(in.Context())

	// metadata
	createReq.RequestContentType = vars.GetFromHeader(in.Header, httperrorutil.HeaderKeyContentType)
	createReq.RequestHeader = func() string {
		b, err := json.Marshal(in.Header)
		if err != nil {
			return err.Error()
		}
		return string(b)
	}()

	createReq.IdentityPhoneNumber = vars.GetFromHeader(in.Header, vars.XAIProxyPhone)
	createReq.IdentityJobNumber = vars.GetFromHeader(in.Header, vars.XAIProxyJobNumber)

	createReq.DingtalkStaffId = vars.GetFromHeader(in.Header, vars.XAIProxyDingTalkStaffID)
	createReq.DingtalkChatType = vars.GetFromHeader(in.Header, vars.XAIProxyChatType)
	createReq.DingtalkChatTitle = vars.GetFromHeader(in.Header, vars.XAIProxyChatTitle)
	createReq.DingtalkChatId = vars.GetFromHeader(in.Header, vars.XAIProxyChatId)

	createReq.BizSource = vars.GetFromHeader(in.Header, vars.XAIProxySource)
	createReq.Username = vars.GetFromHeader(in.Header, vars.XAIProxyUsername, vars.XAIProxyName)
	createReq.Email = vars.GetFromHeader(in.Header, vars.XAIProxyEmail)

	// insert audit into db
	newAudit, err := ctxhelper.MustGetDBClient(in.Context()).AuditClient().CreateWhenReceived(in.Context(), &createReq)
	if err != nil {
		l := ctxhelper.MustGetLoggerBase(in.Context())
		l.Errorf("failed to create audit: %v", err)
	}
	if newAudit != nil {
		ctxhelper.PutAuditID(in.Context(), newAudit.Id)
		ctxhelper.PutAuditSink(in.Context(), types.New(newAudit.Id, ctxhelper.MustGetLoggerBase(in.Context())))
	}
	// add operation_id
	audithelper.Note(in.Context(), "operation_id", in.Method+" "+in.URL.Path)
	// add x-ai-proxy-generated-call-id
	audithelper.Note(in.Context(), "x_ai_proxy_generated_call_id", ctxhelper.MustGetGeneratedCallID(in.Context()))
	return nil
}
