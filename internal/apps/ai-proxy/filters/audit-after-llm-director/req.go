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

package audit_after_llm_director

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/excerptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *Filter) OnActualRequest(ctx context.Context, infor reverseproxy.HttpInfor) {
	// del all X-AI-Proxy-* headers before invoking llm
	for k := range infor.Header() {
		if strings.HasPrefix(strings.ToUpper(k), strings.ToUpper(vars.XAIProxyHeaderPrefix)) {
			infor.Header().Del(k)
		}
	}

	auditRecID, ok := ctxhelper.GetAuditID(ctx)
	if !ok || auditRecID == "" {
		return
	}

	// collect actual llm request info
	updateReq := pb.AuditUpdateRequestAfterLLMDirectorInvoke{
		AuditId:           auditRecID,
		ActualRequestBody: excerptor.ExcerptActualRequestBody(infor.BodyBuffer().String()),
		ActualRequestURL:  infor.URL().String(),
		ActualRequestHeader: func() string {
			b, err := json.Marshal(infor.Header())
			if err != nil {
				return err.Error()
			}
			return string(b)
		}(),
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(ctx)
	dbClient := dao.AuditClient()
	_, err := dbClient.UpdateAfterLLMDirectorInvoke(ctx, &updateReq)
	if err != nil {
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit after llm director invoke, audit id: %s, err: %v", auditRecID, err)
	}
	return
}
