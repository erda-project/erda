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

package openai_v1_models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/sashabaranov/go-openai"

	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/common/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
)

const (
	Name = "openai-v1-models"
)

var (
	_ filter_define.ProxyRequestRewriter = (*Filter)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Filter struct {
}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Filter{}
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// Set context for client handling
	ctx := pr.In.Context()
	richClientHandler := ctxhelper.MustGetRichClientHandler(ctx).(*handler_rich_client.ClientHandler)
	// try set clientId by ak
	client, err := akutil.CheckAkOrToken(ctx, pr.In, ctxhelper.MustGetDBClient(ctx))
	if err != nil {
		return http_error.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get client, err: %v", err))
	}
	if client == nil {
		return http_error.NewHTTPError(http.StatusUnauthorized, "Client not found")
	}
	ctxhelper.PutClient(ctx, client)
	ctxhelper.PutClientId(ctx, client.Id)

	richClient, err := richClientHandler.GetByAccessKeyId(ctx, &richclientpb.GetByClientAccessKeyIdRequest{AccessKeyId: client.AccessKeyId})
	if err != nil {
		return http_error.NewHTTPError(http.StatusInternalServerError, "Failed to get rich client")
	}
	if richClient == nil {
		return http_error.NewHTTPError(http.StatusUnauthorized, "Client not found")
	}
	// convert to openai /v1/models response, see: https://platform.openai.com/docs/api-reference/models/list
	var oaiFormatModels []ExtendedOpenAIModelForList
	for _, m := range richClient.Models {
		oaiFormatModels = append(oaiFormatModels, ExtendedOpenAIModelForList{
			Model: openai.Model{
				ID:        GenerateModelNameWithPublisher(m.Model),
				CreatedAt: m.Model.CreatedAt.Seconds, // seconds
				Object:    "model",                   // always "model"
				OwnedBy:   GetModelPublisher(m.Model),
			},
			Name: GetModelDisplayName(m.Model),
		})
	}
	// construct filter-generated response
	responseBodyBytes, err := json.Marshal(&ModelListResponse{Data: oaiFormatModels})
	if err != nil {
		return http_error.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to marshal response body, err: %v", err))
	}

	resp := &http.Response{
		StatusCode:    http.StatusOK,
		Header:        http.Header{"Content-Type": []string{"application/json"}, "Content-Length": []string{fmt.Sprintf("%d", len(responseBodyBytes))}},
		Body:          io.NopCloser(bytes.NewReader(responseBodyBytes)),
		ContentLength: int64(len(responseBodyBytes)),
		Request:       pr.Out,
	}

	// trigger custom transport: RequestFilterGeneratedResponseTransport
	transports.TriggerRequestFilterGeneratedResponse(pr.Out, resp)

	return nil
}

type ModelListResponse struct {
	Data []ExtendedOpenAIModelForList `json:"data"`
}

type ExtendedOpenAIModelForList struct {
	openai.Model

	Name string `json:"name"`
}
