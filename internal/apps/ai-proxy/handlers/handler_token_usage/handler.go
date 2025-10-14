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

package handler_token_usage

import (
	"context"

	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_i18n/i18n_services"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type TokenUsageHandler struct {
	DAO   dao.DAO
	Cache cachetypes.Manager
}

func (h *TokenUsageHandler) Create(ctx context.Context, req *usagepb.TokenUsageCreateRequest) (*usagepb.TokenUsage, error) {
	return h.DAO.TokenUsageClient().Create(ctx, req)
}

func (h *TokenUsageHandler) Paging(ctx context.Context, req *usagepb.TokenUsagePagingRequest) (*usagepb.TokenUsagePagingResponse, error) {
	if err := enforceSelfScope(ctx, req); err != nil {
		return nil, err
	}
	// limit input page size
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	return h.DAO.TokenUsageClient().Paging(ctx, req)
}

func (h *TokenUsageHandler) Aggregate(ctx context.Context, req *usagepb.TokenUsagePagingRequest) (*usagepb.TokenUsageAggregateResponse, error) {
	if err := enforceSelfScope(ctx, req); err != nil {
		return nil, err
	}

	lang := ""
	if l, ok := ctxhelper.GetAccessLang(ctx); ok {
		lang = l
	}
	locale := i18n_services.GetLocaleFromContext(lang)

	usageRecords, err := h.DAO.TokenUsageClient().Aggregate(ctx, req)
	if err != nil {
		return nil, err
	}
	return h.aggregateTokenUsages(ctx, usageRecords, locale)
}

func enforceSelfScope(ctx context.Context, req *usagepb.TokenUsagePagingRequest) error {
	if ctxhelper.MustGetIsAdmin(ctx) {
		return nil
	}

	clientID := ctxhelper.MustGetClientId(ctx)
	req.ClientId = clientID

	if clientToken, ok := ctxhelper.GetClientToken(ctx); ok && clientToken != nil {
		req.ClientTokenId = clientToken.Id
	}

	return nil
}
