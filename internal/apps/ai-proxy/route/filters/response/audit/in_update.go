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
	"net/http"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (f *AuditResponse) inUpdateEnable(resp *http.Response) bool {
	auditRecID, ok := ctxhelper.GetAuditID(resp.Request.Context())
	return ok && auditRecID != ""
}

func (f *AuditResponse) inUpdateOnHeaders(resp *http.Response) error {
	return nil
}

func (f *AuditResponse) inUpdateOnBodyChunk(resp *http.Response, chunk []byte) (out []byte, err error) {
	if !f.inUpdateEnable(resp) {
		return chunk, nil
	}
	if f.inUpdate.firstResponseAt.IsZero() {
		f.inUpdate.firstResponseAt = time.Now()
	}
	f.inUpdate.allChunks = append(f.inUpdate.allChunks, chunk...)
	return chunk, nil
}

func (f *AuditResponse) inUpdateOnComplete(resp *http.Response) (out []byte, err error) {
	if !f.inUpdateEnable(resp) {
		return nil, nil
	}
	auditRecID, _ := ctxhelper.GetAuditID(resp.Request.Context())
	// collect actual llm response info
	updateReq := pb.AuditUpdateRequestAfterLLMResponse{
		AuditId:    auditRecID,
		ResponseAt: timestamppb.New(f.inUpdate.firstResponseAt),
		Status:     int32(resp.StatusCode),
		//ActualResponseBody: string(f.inUpdate.allChunks), // TODO not store raw body anymore, but store parsed content
		ActualResponseHeader: func() string {
			if actualHeader, err := json.Marshal(resp.Header); err != nil {
				return err.Error()
			} else {
				return string(actualHeader)
			}
		}(),
		ResponseContentType: resp.Header.Get(httputil.HeaderKeyContentType),
		ResponseStreamDoneAt: func() *timestamppb.Timestamp {
			if ctxhelper.GetIsStream(resp.Request.Context()) {
				return timestamppb.New(time.Now())
			}
			return nil
		}(),
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(resp.Request.Context())
	dbClient := dao.AuditClient()
	_, err = dbClient.UpdateAfterLLMResponse(resp.Request.Context(), &updateReq)
	if err != nil {
		l := ctxhelper.MustGetLogger(resp.Request.Context())
		l.Errorf("failed to update audit after llm response, audit id: %s, err: %v", auditRecID, err)
	}
	return nil, nil
}
