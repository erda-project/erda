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

//go:generate go run ./generate/generate_by_keys.go

package ctxhelper

import (
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
)

type (
	// Keys migrated from vars package
	mapKeyClient          struct{ *clientpb.Client }
	mapKeyModel           struct{ *modelpb.Model }
	mapKeyServiceProvider struct {
		*serviceproviderpb.ServiceProvider
	}
	mapKeyPromptTemplate  struct{ *promptpb.Prompt }
	mapKeySession         struct{ *sessionpb.Session }
	mapKeyClientToken     struct{ *clienttokenpb.ClientToken }
	mapKeyMessageGroup    struct{ message.Group }
	mapKeyIsStream        struct{ bool }
	mapKeyAuditID         struct{ string }
	mapKeyAuditSink       struct{ types.Sink }
	mapKeyRequestID       struct{ string }
	mapKeyGeneratedCallID struct{ string }
	mapKeyLogger          struct{ logs.Logger }
	mapKeyLoggerBase      struct{ logs.Logger }

	mapKeyMcpInfo struct {
		McpInfo
	}

	// Keys for response processing
	mapKeyRespBodyChunkSplitter struct {
		filter_define.RespBodyChunkSplitter
	}
	mapKeyResponseContentEncoding struct{ string }

	// Keys for filter-generated responses
	mapKeyRequestFilterGeneratedResponse struct{ *http.Response }

	// Keys for reverse proxy
	mapKeyReverseProxyRequestRewriteError struct{ *ReverseProxyFilterError }
	mapKeyReverseProxyResponseModifyError struct{ *ReverseProxyFilterError }
	mapKeyReverseProxyRequestInSnapshot   struct{ *http.Request }
	mapKeyReverseProxyRequestOutSnapshot  struct{ *http.Request }

	mapKeyReverseProxyWholeHandledResponseBodyStr struct{ string }

	// Keys for migrated context keys
	mapKeyDBClient     struct{ dao.DAO }
	mapKeyPathMatcher  struct{ *path_matcher.PathMatcher }
	mapKeyCacheManager struct{ any }

	// Additional context keys migrated from vars
	mapKeyIsAdmin         struct{ bool }
	mapKeyClientId        struct{ string }
	mapKeyAccessLang      struct{ string }
	mapKeyAIProxyHandlers struct{ any }

	mapKeyRequestBodyTransformChanges     struct{ any }
	mapKeyRequestThinkingTransformChanges struct{ any }

	mapKeyRequestBeginAt struct{ time.Time }
)

// KeysWithCustomMustGet defines keys with custom MustGet implementations (should not generate default MustGet)
var KeysWithCustomMustGet = map[any]bool{
	mapKeyIsStream{}: true,
	mapKeyIsAdmin{}:  true,
}
