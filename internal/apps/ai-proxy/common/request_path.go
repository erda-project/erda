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

package common

import (
	"context"

	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	RequestPathPrefixV1ChatCompletions = "/v1/chat/completions"
	RequestPathPrefixV1Completions     = "/v1/completions"
	RequestPathPrefixV1Images          = "/v1/images"
	RequestPathPrefixV1Audio           = "/v1/audio"
	RequestPathPrefixV1Embeddings      = "/v1/embeddings"
	RequestPathPrefixV1Moderations     = "/v1/moderations"

	RequestPathPrefixV1Files = "/v1/files"

	RequestPathPrefixV1Assistants    = "/v1/assistants"
	RequestPathPrefixV1Responses     = "/v1/responses"
	RequestPathPrefixV1Threads       = "/v1/threads"
	RequestPathV1ThreadCreateMessage = "/v1/threads/{thread_id}/messages"
)

func GetRequestRoutePath(ctx context.Context) string {
	return ctx.Value(reverseproxy.CtxKeyPath{}).(string)
}
