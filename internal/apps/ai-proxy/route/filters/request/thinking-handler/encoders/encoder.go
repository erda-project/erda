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

package encoders

import (
	"context"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// CommonThinkingEncoder defines the interface for encoding CommonThinking to provider-specific format
type CommonThinkingEncoder interface {
	// CanEncode checks if this encoder can handle the given context
	CanEncode(ctx context.Context) bool

	// Encode encodes CommonThinking to provider-specific format in the request body
	Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error)

	// GetPriority returns the priority of this encoder (lower number = higher priority)
	GetPriority() int

	// GetName returns the name of this encoder for logging and debugging
	GetName() string
}
