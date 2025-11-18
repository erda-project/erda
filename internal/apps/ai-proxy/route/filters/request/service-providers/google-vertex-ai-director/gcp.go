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

package google_vertex_ai_director

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/google"

	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func getGCPProjectId(ctx context.Context, sp *providerpb.ServiceProvider) (string, error) {
	// read content of credential.json（service_account）
	keyFileContent := sp.TemplateParams["service-account-key-file-content"]
	if keyFileContent == "" {
		return "", fmt.Errorf("service-account-key-file-content not found in template params")
	}
	// parse content as json
	var obj map[string]any
	if err := json.Unmarshal([]byte(keyFileContent), &obj); err != nil {
		return "", fmt.Errorf("failed to parse json key file content: %v", err)
	}
	v, ok := obj["project_id"]
	if !ok {
		return "", fmt.Errorf("project_id not found in json key file content")
	}
	pid, ok := v.(string)
	if !ok || pid == "" {
		return "", fmt.Errorf("invalid project_id in json key file content")
	}
	return pid, nil
}

func getGCPAccessToken(ctx context.Context, sp *providerpb.ServiceProvider) (string, error) {
	// 1. read content of credential.json（service_account）
	keyFileContent := sp.TemplateParams["service-account-key-file-content"]
	if keyFileContent == "" {
		return "", fmt.Errorf("service-account-key-file-content not found in template params")
	}
	// 2. use SA JSON to create TokenSource（cloud-platform scope）
	const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
	creds, err := google.CredentialsFromJSON(ctx, []byte(keyFileContent), cloudPlatformScope)
	if err != nil {
		// print detailed file content
		ctxhelper.MustGetLogger(ctx).Errorf("failed to parse credentials from key file content,err: %v, content: %s", err, keyFileContent)
		// return fixed error content without leaking sensitive info
		return "", fmt.Errorf("failed to parse credentials from key file content")
	}
	ts := creds.TokenSource

	// 3. get a short-term access token
	token, err := ts.Token()
	if err != nil {
		ctxhelper.MustGetLogger(ctx).Errorf("failed to get access token from gcp: %v", err)
		return "", fmt.Errorf("failed to get access token from gcp")
	}
	accessToken := token.AccessToken
	return accessToken, nil
}
