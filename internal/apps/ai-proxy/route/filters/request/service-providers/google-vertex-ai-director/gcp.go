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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
)

var (
	// gcpProjectIDCache caches parsed project_id values keyed by the service-account key file content.
	gcpProjectIDCache sync.Map // map[string]string

	// gcpTokenSourceCache caches oauth2.TokenSource instances keyed by service-account key content and scope.
	gcpTokenSourceCache sync.Map // map[string]oauth2.TokenSource
)

func cacheKeyFromContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func getGCPProjectId(ctx context.Context, sp *providerpb.ServiceProvider) (string, error) {
	// read content of credential.json (service_account)
	keyFileContent := sp.TemplateParams["service-account-key-file-content"]
	if keyFileContent == "" {
		return "", fmt.Errorf("service-account-key-file-content not found in template params")
	}
	cacheKey := cacheKeyFromContent(keyFileContent)

	// fast path: try cache first
	if v, ok := gcpProjectIDCache.Load(cacheKey); ok {
		if pid, ok := v.(string); ok && pid != "" {
			return pid, nil
		}
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

	// store in cache for subsequent calls
	gcpProjectIDCache.Store(cacheKey, pid)

	return pid, nil
}

func getGCPAccessToken(ctx context.Context, sp *providerpb.ServiceProvider) (string, error) {
	// 1. read content of credential.json (service_account)
	keyFileContent := sp.TemplateParams["service-account-key-file-content"]
	if keyFileContent == "" {
		return "", fmt.Errorf("service-account-key-file-content not found in template params")
	}

	// Build a proxy-aware HTTP client for oauth2 token exchange
	oauthHTTPClient := &http.Client{Transport: transports.BaseTransport}
	// Inject the client so oauth2 and google auth flows respect the proxy
	ctx = context.WithValue(ctx, oauth2.HTTPClient, oauthHTTPClient)

	// 2. use SA JSON to create or reuse a TokenSource (cloud-platform scope)
	const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	cacheKey := cacheKeyFromContent(keyFileContent + "|" + cloudPlatformScope)

	var ts oauth2.TokenSource
	if v, ok := gcpTokenSourceCache.Load(cacheKey); ok {
		if cached, ok := v.(oauth2.TokenSource); ok && cached != nil {
			ts = cached
		}
	}

	if ts == nil {
		creds, err := google.CredentialsFromJSON(ctx, []byte(keyFileContent), cloudPlatformScope)
		if err != nil {
			// print detailed file content
			ctxhelper.MustGetLogger(ctx).Errorf("failed to parse credentials from key file content,err: %v, content: %s", err, keyFileContent)
			// return fixed error content without leaking sensitive info
			return "", fmt.Errorf("failed to parse credentials from key file content")
		}
		ts = creds.TokenSource
		gcpTokenSourceCache.Store(cacheKey, ts)
	}

	// 3. get a short-term access token
	token, err := ts.Token()
	if err != nil {
		ctxhelper.MustGetLogger(ctx).Errorf("failed to get access token from gcp: %v", err)
		return "", fmt.Errorf("failed to get access token from gcp")
	}
	accessToken := token.AccessToken
	return accessToken, nil
}
