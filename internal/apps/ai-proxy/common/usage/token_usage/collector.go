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

package token_usage

import (
	"context"
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"

	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage/estimators"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage/extractors"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage/proxy_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

type usageCollector struct {
	dbClient dao.DAO
}

var defaultCollector *usageCollector

func InitUsageCollector(dao dao.DAO) {
	defaultCollector = &usageCollector{dbClient: dao}
}

func Collect(resp *http.Response) {
	if !canCollect(resp) {
		return
	}

	ctx := resp.Request.Context()

	createReq := usagepb.TokenUsageCreateRequest{
		CallId:     ctxhelper.MustGetGeneratedCallID(ctx),
		XRequestId: ctxhelper.MustGetRequestID(ctx),
		ClientId:   ctxhelper.MustGetClientId(ctx),
		ClientTokenId: func() string {
			clientToken, ok := ctxhelper.GetClientToken(ctx)
			if !ok {
				return ""
			}
			return clientToken.Id
		}(),
		ProviderId: ctxhelper.MustGetServiceProvider(ctx).Id,
		ModelId:    ctxhelper.MustGetModel(ctx).Id,
		CreatedAt:  timestamppb.New(ctxhelper.MustGetRequestBeginAt(ctx)),
		Metadata:   nil,
	}

	// fulfill token related fields
	calculateTokens(resp, &createReq)

	if _, err := defaultCollector.dbClient.TokenUsageClient().Create(context.Background(), &createReq); err != nil {
		ctxhelper.MustGetLogger(ctx).Errorf("fail to create token usage record: %v", err)
	}
}

func calculateTokens(resp *http.Response, createReq *usagepb.TokenUsageCreateRequest) {
	ctx := resp.Request.Context()
	proxyType := proxy_types.DetermineProxyType(ctx)

	extractor, ok := extractors.TryGetExtractorByProxyType(proxyType)
	if ok {
		if success := extractor.TryExtract(resp, createReq); success {
			createReq.IsEstimated = false
			return
		}
	}
	estimator := estimators.MustGetEstimatorByProxyType(proxyType)
	if success := estimator.Estimate(resp, createReq); success {
		createReq.IsEstimated = true
		return
	}
	// calculate failed
	ctxhelper.MustGetLogger(ctx).Warnf("fail to estimate token usage record for proxy type: %s", proxyType)
}

func canCollect(resp *http.Response) bool {
	ctx := resp.Request.Context()

	// model
	if _, ok := ctxhelper.GetModel(ctx); !ok {
		return false
	}

	return true
}
