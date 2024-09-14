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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/decompress"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *Filter) OnResponseChunkImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (signal reverseproxy.Signal, err error) {
	if f.firstResponseAt.IsZero() {
		f.firstResponseAt = time.Now()
	}
	return reverseproxy.Continue, nil
}

func (f *Filter) OnResponseEOFImmutable(ctx context.Context, infor reverseproxy.HttpInfor, copiedChunk []byte) (err error) {
	auditRecID, ok := ctxhelper.GetAuditID(ctx)
	if !ok || auditRecID == "" {
		return nil
	}
	// collect actual llm response info
	updateReq := pb.AuditUpdateRequestAfterLLMResponse{
		AuditId:            auditRecID,
		ResponseAt:         timestamppb.New(f.firstResponseAt),
		Status:             int32(infor.StatusCode()),
		ActualResponseBody: string(decompress.TryDecompressBody(infor.Header(), f.DefaultResponseFilter.Buffer)),
		ActualResponseHeader: func() string {
			if actualHeader, err := json.Marshal(infor.Header()); err != nil {
				return err.Error()
			} else {
				return string(actualHeader)
			}
		}(),
		ResponseContentType: infor.Header().Get(httputil.HeaderKeyContentType),
		ResponseStreamDoneAt: func() *timestamppb.Timestamp {
			if ctxhelper.GetIsStream(ctx) {
				return timestamppb.New(time.Now())
			}
			return nil
		}(),
	}

	// update audit into db
	dao := ctxhelper.MustGetDBClient(ctx)
	dbClient := dao.AuditClient()
	_, err = dbClient.UpdateAfterLLMResponse(ctx, &updateReq)
	if err != nil {
		l := ctxhelper.GetLogger(ctx)
		l.Errorf("failed to update audit after llm response, audit id: %s, err: %v", auditRecID, err)
	}
	return nil
}
