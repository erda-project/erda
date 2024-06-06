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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"

	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/common/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers/handler_rich_client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "openai-v1-models"
)

var (
	_ reverseproxy.RequestFilter = (*Filter)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type Filter struct {
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &Filter{}, nil
}

func (f *Filter) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	var (
		richClientHandler = ctx.Value(vars.CtxKeyRichClientHandler{}).(*handler_rich_client.ClientHandler)
	)
	// try set clientId by ak
	client, err := akutil.CheckAkOrToken(ctx, infor.Request(), ctxhelper.MustGetDBClient(ctx))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get client, err: %v", err), http.StatusInternalServerError)
		return reverseproxy.Intercept, nil
	}
	if client == nil {
		http.Error(w, "Client not found", http.StatusUnauthorized)
		return reverseproxy.Intercept, nil
	}
	ctx = context.WithValue(ctx, vars.CtxKeyClient{}, client)
	ctx = context.WithValue(ctx, vars.CtxKeyClientId{}, client.Id)

	richClient, err := richClientHandler.GetByAccessKeyId(ctx, &richclientpb.GetByClientAccessKeyIdRequest{AccessKeyId: client.AccessKeyId})
	if err != nil {
		http.Error(w, "Failed to get rich client", http.StatusInternalServerError)
		return reverseproxy.Intercept, nil
	}
	// convert to openai /v1/models response, see: https://platform.openai.com/docs/api-reference/models/list
	var oaiFormatModels openai.ModelsList
	for _, m := range richClient.Models {
		oaiFormatModels.Models = append(oaiFormatModels.Models, openai.Model{
			ID:        GenerateModelDisplayName(m),
			CreatedAt: m.Model.CreatedAt.Seconds, // seconds
			Object:    "model",                   // always "model"
			OwnedBy:   m.Provider.Name,
		})
	}
	// write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(oaiFormatModels); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return reverseproxy.Intercept, nil
	}
	return reverseproxy.Continue, nil
}
