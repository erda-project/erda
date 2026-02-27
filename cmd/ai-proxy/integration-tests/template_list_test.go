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

package integration_tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

type templateListResp struct {
	Success bool `json:"success"`
	Data    struct {
		Total int64            `json:"total"`
		List  []map[string]any `json:"list"`
	} `json:"data"`
	Message string `json:"message"`
}

func TestTemplateListAuthAndNoAuth(t *testing.T) {
	client := common.NewClient()
	cfg := config.Get()

	cases := []struct {
		name string
		path string
	}{
		{name: "ModelTemplates", path: "/api/ai-proxy/templates/types/model"},
		{name: "ProviderTemplates", path: "/api/ai-proxy/templates/types/service-provider"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"_NoAuth", func(t *testing.T) {
			resp := requestTemplateList(t, client, tc.path, true)
			data := decodeTemplateList(t, resp)
			if data.Data.Total <= 0 || len(data.Data.List) == 0 {
				t.Fatalf("expected non-empty template list, total=%d len=%d", data.Data.Total, len(data.Data.List))
			}
		})

		t.Run(tc.name+"_Auth", func(t *testing.T) {
			if cfg.Token == "" {
				t.Skip("No token configured, skipping auth case")
			}
			resp := requestTemplateList(t, client, tc.path, false)
			data := decodeTemplateList(t, resp)
			if data.Data.Total <= 0 || len(data.Data.List) == 0 {
				t.Fatalf("expected non-empty template list, total=%d len=%d", data.Data.Total, len(data.Data.List))
			}
		})
	}
}

func TestTemplateListNoAuthNoSensitiveInfo(t *testing.T) {
	client := common.NewClient()
	// Use a fake UUID-format clientId to detect leakage in response body.
	fakeClientID := "11111111-2222-3333-4444-555555555555"
	query := fmt.Sprintf("?checkInstance=true&clientId=%s&showDeprecated=true&renderTemplate=true", fakeClientID)

	cases := []struct {
		name                 string
		path                 string
		checkDeprecatedField bool
	}{
		{name: "ModelTemplates", path: "/api/ai-proxy/templates/types/model", checkDeprecatedField: true},
		{name: "ProviderTemplates", path: "/api/ai-proxy/templates/types/service-provider", checkDeprecatedField: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp := requestTemplateList(t, client, tc.path+query, true)
			data := decodeTemplateList(t, resp)

			if hasAnyField(data.Data.List, "instanceCount") || hasAnyField(data.Data.List, "enabledInstanceCount") {
				t.Fatalf("no-auth response should not contain instance count fields")
			}

			if tc.checkDeprecatedField && hasDeprecatedTrue(data.Data.List) {
				t.Fatalf("no-auth response should not include deprecated templates")
			}

			if strings.Contains(string(resp.Body), fakeClientID) {
				t.Fatalf("no-auth response should not leak user-provided clientId")
			}
		})
	}
}

func TestTemplateListAuthCheckInstanceVisible(t *testing.T) {
	cfg := config.Get()
	if cfg.Token == "" {
		t.Skip("No token configured, skipping auth case")
	}

	client := common.NewClient()
	cases := []struct {
		name string
		path string
	}{
		{name: "ModelTemplates", path: "/api/ai-proxy/templates/types/model?checkInstance=true"},
		{name: "ProviderTemplates", path: "/api/ai-proxy/templates/types/service-provider?checkInstance=true"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp := requestTemplateList(t, client, tc.path, false)
			data := decodeTemplateList(t, resp)
			if len(data.Data.List) == 0 {
				t.Fatalf("expected non-empty template list")
			}
			if !hasAnyField(data.Data.List, "instanceCount") || !hasAnyField(data.Data.List, "enabledInstanceCount") {
				t.Fatalf("auth response with checkInstance=true should contain instance count fields")
			}
		})
	}
}

func requestTemplateList(t *testing.T, client *common.Client, path string, noAuth bool) *common.APIResponse {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	var resp *common.APIResponse
	if noAuth {
		resp = client.GetWithHeaders(ctx, path, map[string]string{"Authorization": ""})
	} else {
		resp = client.Get(ctx, path)
	}

	if resp.Error != nil {
		t.Fatalf("request failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}
	return resp
}

func decodeTemplateList(t *testing.T, resp *common.APIResponse) *templateListResp {
	t.Helper()
	var out templateListResp
	if err := resp.GetJSON(&out); err != nil {
		t.Fatalf("failed to parse template list response: %v", err)
	}
	if !out.Success {
		t.Fatalf("API returned success=false: %s", out.Message)
	}
	return &out
}

func hasAnyField(list []map[string]any, field string) bool {
	for _, item := range list {
		if _, ok := item[field]; ok {
			return true
		}
	}
	return false
}

func hasDeprecatedTrue(list []map[string]any) bool {
	for _, item := range list {
		v, ok := item["deprecated"]
		if !ok {
			continue
		}
		b, ok := v.(bool)
		if ok && b {
			return true
		}
	}
	return false
}
