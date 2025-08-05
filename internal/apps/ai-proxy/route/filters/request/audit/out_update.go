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

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/excerptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/audit/audit_util"
)

func (f *Audit) outUpdate(out *http.Request) error {
	auditRecID, ok := ctxhelper.GetAuditID(out.Context())
	if !ok || auditRecID == "" {
		return nil
	}

	// collect actual llm request info - decide whether to save body based on content-type
	contentType := out.Header.Get("Content-Type")
	shouldSaveBody := audit_util.ShouldDumpBody(contentType)

	var actualRequestBody string
	if shouldSaveBody {
		bodyCopy, err := body_util.SmartCloneBody(&out.Body, body_util.MaxSample)
		if err != nil {
			return fmt.Errorf("fail to clone body: %w", err)
		}
		bodyCopyBytes, err := io.ReadAll(bodyCopy)
		if err != nil {
			return fmt.Errorf("fail to read cloned body: %w", err)
		}
		actualRequestBody = excerptor.ExcerptActualRequestBody(string(bodyCopyBytes))
	} else {
		// For binary content, only save a descriptive message
		actualRequestBody = fmt.Sprintf("[Binary content omitted - content-type: %s]", contentType)
	}

	updateReq := pb.AuditUpdateRequestAfterLLMDirectorInvoke{
		AuditId:           auditRecID,
		ActualRequestBody: actualRequestBody,
		ActualRequestURL:  out.URL.String(),
		ActualRequestHeader: func() string {
			b, marshalErr := json.Marshal(out.Header)
			if marshalErr != nil {
				return marshalErr.Error()
			}
			return string(b)
		}(),
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(out.Context())
	dbClient := dao.AuditClient()
	if _, updateErr := dbClient.UpdateAfterLLMDirectorInvoke(out.Context(), &updateReq); updateErr != nil {
		ctxhelper.MustGetLogger(out.Context()).Warnf("fail to update audit after llm director invoke: %v", updateErr)
	}

	return nil
}
