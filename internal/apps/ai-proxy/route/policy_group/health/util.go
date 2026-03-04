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

package health

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func buildProbeRequest(apiType APIType) (path string, body []byte, ok bool) {
	switch apiType {
	case APITypeChatCompletions:
		return vars.RequestPathPrefixV1ChatCompletions, []byte(`{"model":"health-probe","messages":[{"role":"user","content":"hello"}],"stream":false}`), true
	case APITypeResponses:
		return vars.RequestPathPrefixV1Responses, []byte(`{"model":"health-probe","input":"hello","stream":false}`), true
	default:
		return "", nil, false
	}
}

func isAPITypeProbeSupported(apiType APIType) bool {
	_, _, ok := buildProbeRequest(apiType)
	return ok
}

func buildProbeHeadersFromRequestMeta(meta policygroup.RequestMeta) http.Header {
	headers := make(http.Header)
	for metaKey, metaValue := range meta.Keys {
		if !strings.HasPrefix(strings.ToLower(metaKey), common_types.StickyKeyPrefixFromReqHeader) {
			continue
		}
		headerKey := strings.TrimSpace(metaKey[len(common_types.StickyKeyPrefixFromReqHeader):])
		if headerKey == "" {
			continue
		}
		if strings.TrimSpace(metaValue) == "" {
			continue
		}
		headers.Set(http.CanonicalHeaderKey(headerKey), metaValue)
	}
	return headers
}

func cloneHeaders(headers http.Header) http.Header {
	if len(headers) == 0 {
		return http.Header{}
	}
	cloned := make(http.Header, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func withJitter(backoff time.Duration) time.Duration {
	if backoff <= 0 {
		return 0
	}
	jitterRange := backoff / 4
	if jitterRange <= 0 {
		return backoff
	}
	delta := time.Duration(rand.Int63n(int64(jitterRange)))
	return backoff + delta
}

func extractCallID(headers http.Header) string {
	if len(headers) == 0 {
		return ""
	}
	if v := strings.TrimSpace(headers.Get(vars.XAIProxyGeneratedCallId)); v != "" {
		return v
	}
	if v := strings.TrimSpace(headers.Get(vars.XRequestId)); v != "" {
		return v
	}
	return ""
}

func extractRequestID(headers http.Header) string {
	if len(headers) == 0 {
		return ""
	}
	return strings.TrimSpace(headers.Get(vars.XRequestId))
}

func tryPutModelMarkUnhealthyInstanceID(ctx context.Context, instanceID string) {
	if ctx == nil || instanceID == "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("instance_id", instanceID).Debugf("skip model unhealthy ctx mark: %v", r)
		}
	}()
	ctxhelper.PutModelMarkUnhealthyInstanceID(ctx, instanceID)
}
