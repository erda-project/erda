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

package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/excerptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/audit/audit_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

func (f *Audit) requestInCreateAudit(in *http.Request) error {
	var createReq pb.AuditCreateRequestWhenReceived
	createReq.RequestAt = timestamppb.New(time.Now())
	createReq.AuthKey = vars.TrimBearer(in.Header.Get(vars.TrimBearer(httperrorutil.HeaderKeyAuthorization)))
	// request body - decide whether to save body based on content-type
	contentType := in.Header.Get("Content-Type")
	shouldSaveBody := audit_util.ShouldDumpBody(contentType)

	if shouldSaveBody {
		bodyCopy, err := body_util.SmartCloneBody(&in.Body, body_util.MaxSample)
		if err != nil {
			return fmt.Errorf("failed to clone request body: %w", err)
		}
		bodyCopyBytes, err := io.ReadAll(bodyCopy)
		if err != nil {
			return fmt.Errorf("failed to read cloned request body: %w", err)
		}
		createReq.RequestBody = excerptor.ExcerptActualRequestBody(string(bodyCopyBytes))
	} else {
		// For binary content, only save a descriptive message
		createReq.RequestBody = fmt.Sprintf("[Binary content omitted - content-type: %s]", contentType)
	}
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
		l := ctxhelper.MustGetLogger(in.Context())
		l.Errorf("failed to create audit: %v", err)
	}
	if newAudit != nil {
		ctxhelper.PutAuditID(in.Context(), newAudit.Id)
	}
	return nil
}
