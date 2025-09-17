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

package handler_audit

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/audit/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/pkg/desensitize"
)

type AuditHandler struct {
	DAO dao.DAO
}

func (a *AuditHandler) Paging(ctx context.Context, req *pb.AuditPagingRequest) (*pb.AuditPagingResponse, error) {
	isAdmin, err := checkAndFillAuth(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := requireXRequestIdForNonAdmin(ctx, req); err != nil {
		return nil, err
	}

	pagingResult, err := a.DAO.AuditClient().Paging(ctx, req)
	if err != nil {
		return nil, err
	}

	// desensitize
	if !isAdmin {
		for _, audit := range pagingResult.List {
			audit.AuthKey = ""
			audit.ActualRequestBody = "[omitted due to sensitive data]"
			audit.Username = desensitize.Name(audit.Username)
			audit.Email = desensitize.Email(audit.Email)
			// filter metadata
			if audit.Metadata != nil {
				for k := range audit.Metadata.Public {
					switch k {
					case "request_begin_at":
					case "response_chunk_begin_at", "response_chunk_done_at":
					default:
						delete(audit.Metadata.Public, k)
					}
				}
				audit.Metadata.Secret = nil
			}
		}
	}
	return pagingResult, nil
}

// checkAndFillAuth checks admin or client identity and fills req.AuthKey for non-admin callers.
// It also enforces non-admin callers to provide an XRequestId to narrow down search scope.
// Returns whether the caller is admin.
func checkAndFillAuth(ctx context.Context, req *pb.AuditPagingRequest) (bool, error) {
	isAdmin := ctxhelper.MustGetIsAdmin(ctx)
	if !isAdmin {
		if clientToken, ok := ctxhelper.GetClientToken(ctx); ok {
			// prefer client token
			req.AuthKey = clientToken.Token
		} else {
			client := ctxhelper.MustGetClient(ctx)
			req.AuthKey = client.AccessKeyId
		}
	}
	return isAdmin, nil
}

// requireXRequestIdForNonAdmin enforces that non-admin calls provide x_request_id to narrow down the search scope.
func requireXRequestIdForNonAdmin(ctx context.Context, req *pb.AuditPagingRequest) error {
	if !ctxhelper.MustGetIsAdmin(ctx) {
		if req.XRequestId == "" {
			return fmt.Errorf("missing query param: x_request_id")
		}
	}
	return nil
}
