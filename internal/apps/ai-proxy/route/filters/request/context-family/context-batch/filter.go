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

package context

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

const (
	Name = "context-batch"
)

type CreateBatchRequest struct {
	InputFileID      string
	Endpoint         string
	CompletionWindow string
}

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type Context struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	path := ctxhelper.MustGetPathMatcher(pr.In.Context()).Pattern

	switch {
	case path == vars.RequestPathPrefixV1Batches && pr.In.Method == http.MethodPost:
		return onCreateBatch(pr)
	case path == vars.RequestPathPrefixV1Batches && pr.In.Method == http.MethodGet:
		audithelper.Note(pr.In.Context(), "prompt", "batch list request")
	case path == vars.RequestPathPrefixV1BatchesByID && pr.In.Method == http.MethodGet:
		audithelper.Note(pr.In.Context(), "prompt", fmt.Sprintf("batch retrieve request, batch_id: %s", getBatchID(pr)))
	case path == vars.RequestPathPrefixV1BatchesCancel && pr.In.Method == http.MethodPost:
		audithelper.Note(pr.In.Context(), "prompt", fmt.Sprintf("batch cancel request, batch_id: %s", getBatchID(pr)))
	}

	return nil
}

func onCreateBatch(pr *httputil.ProxyRequest) error {
	bodyCopy, err := body_util.SmartCloneBody(&pr.In.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}
	if bodyCopy.Size() == 0 {
		return fmt.Errorf("request body is empty")
	}

	var raw map[string]any
	if err := json.NewDecoder(bodyCopy).Decode(&raw); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	req, err := parseCreateBatchRequest(raw)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(
		"batch endpoint: %s, completion_window: %s, input_file_id: %s",
		req.Endpoint,
		req.CompletionWindow,
		req.InputFileID,
	)
	audithelper.Note(pr.In.Context(), "prompt", prompt)
	return nil
}

func parseCreateBatchRequest(raw map[string]any) (*CreateBatchRequest, error) {
	inputFileID, err := requiredString(raw, "input_file_id")
	if err != nil {
		return nil, err
	}
	endpoint, err := requiredString(raw, "endpoint")
	if err != nil {
		return nil, err
	}
	completionWindow, err := requiredString(raw, "completion_window")
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(endpoint, "/") {
		return nil, fmt.Errorf("endpoint must start with '/'")
	}

	return &CreateBatchRequest{
		InputFileID:      inputFileID,
		Endpoint:         endpoint,
		CompletionWindow: completionWindow,
	}, nil
}

func requiredString(raw map[string]any, key string) (string, error) {
	v, ok := raw[key]
	if !ok {
		return "", fmt.Errorf("%s is required", key)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return strings.TrimSpace(s), nil
}

func getBatchID(pr *httputil.ProxyRequest) string {
	if batchID, ok := ctxhelper.GetPathParam(pr.In.Context(), "batch_id"); ok && strings.TrimSpace(batchID) != "" {
		return batchID
	}
	return "unknown"
}
