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

	"github.com/erda-project/erda-infra/base/logs"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	promptpb "github.com/erda-project/erda-proto-go/apps/aiproxy/prompt/pb"
	sessionpb "github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
)

type (
	// Keys migrated from vars package
	mapKeyClient             struct{ *clientpb.Client }
	mapKeyModel              struct{ *modelpb.Model }
	mapKeyModelProvider      struct{ *modelproviderpb.ModelProvider }
	mapKeyPromptTemplate     struct{ *promptpb.Prompt }
	mapKeySession            struct{ *sessionpb.Session }
	mapKeyClientToken        struct{ *clienttokenpb.ClientToken }
	mapKeyMessageGroup       struct{ message.Group }
	mapKeyUserPrompt         struct{ string }
	mapKeyIsStream           struct{ bool }
	mapKeyAuditID            struct{ string }
	mapKeyRequestID          struct{ string }
	mapKeyGeneratedCallID    struct{ string }
	mapKeyResponseChunkIndex struct{ int }
	mapKeyAudioInfo          struct{ AudioInfo }
	mapKeyImageInfo          struct{ ImageInfo }
	mapKeyLogger             struct{ logs.Logger }

	mapKeyMcpInfo struct {
		McpInfo
	}

	// Keys for response processing
	mapKeyRespBodyChunkSplitter struct {
		filter_define.RespBodyChunkSplitter
	}
	mapKeyResponseContentEncoding struct{ string }
	mapKeyResponseModifierError   struct{ error }

	// Keys for filter-generated responses
	mapKeyRequestFilterGeneratedResponse struct{ *http.Response }

	// Keys for reverse proxy
	mapKeyReverseProxyAtRewriteStage struct{ error }

	// Keys for migrated context keys
	mapKeyDBClient    struct{ dao.DAO }
	mapKeyPathMatcher struct{ *path_matcher.PathMatcher }

	// Additional context keys migrated from vars
	mapKeyIsAdmin           struct{ bool }
	mapKeyClientId          struct{ string }
	mapKeyAccessLang        struct{ string }
	mapKeyRichClientHandler struct{ any }

	mapKeyRequestBodyTransformChanges     struct{ any }
	mapKeyRequestThinkingTransformChanges struct{ any }
)

// KeysWithCustomMustGet defines keys with custom MustGet implementations (should not generate default MustGet)
var KeysWithCustomMustGet = map[any]bool{
	mapKeyLogger{}:             true,
	mapKeyResponseChunkIndex{}: true,
	mapKeyIsStream{}:           true,
}
