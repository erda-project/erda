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

package anthropic_compatible_director

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/anthropic_official"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/aws_bedrock"
)

var (
	_ filter_define.ProxyResponseModifier = (*AnthropicCompatibleDirectorResponse)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &AnthropicCompatibleDirectorResponse{
		BedrockDirector:   aws_bedrock.NewDirector(),
		AnthropicDirector: anthropic_official.NewDirector(),
	}
}

func init() {
	filter_define.RegisterFilterCreator("anthropic-compatible-director", ResponseModifierCreator)
}

type AnthropicCompatibleDirectorResponse struct {
	BedrockDirector   *aws_bedrock.BedrockDirector
	AnthropicDirector *anthropic_official.AnthropicDirector
}

func (f *AnthropicCompatibleDirectorResponse) Enable(ctx context.Context) bool {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleAnthropicCompatible))
}

func (f *AnthropicCompatibleDirectorResponse) OnHeaders(resp *http.Response) error {
	if !f.Enable(resp.Request.Context()) {
		return nil
	}
	return nil
}

func (f *AnthropicCompatibleDirectorResponse) OnBodyChunk(resp *http.Response, chunk []byte) ([]byte, error) {
	if !f.Enable(resp.Request.Context()) {
		return chunk, nil
	}
	if resp.StatusCode != http.StatusOK {
		return chunk, nil // do not process non-200 responses
	}
	apiSegment := api_segment_getter.GetAPISegment(resp.Request.Context())
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		return f.BedrockDirector.OnBodyChunk(resp, chunk)
	case strings.ToLower(string(anthropic_official.APIVendor)):
		return f.AnthropicDirector.OnBodyChunk(resp, chunk)
	default:
		return nil, fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}
}

func (f *AnthropicCompatibleDirectorResponse) OnComplete(resp *http.Response) ([]byte, error) {
	if !f.Enable(resp.Request.Context()) {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	apiSegment := api_segment_getter.GetAPISegment(resp.Request.Context())
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		return f.BedrockDirector.OnComplete(resp)
	case strings.ToLower(string(anthropic_official.APIVendor)):
		return f.AnthropicDirector.OnComplete(resp)
	default:
		return nil, fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}
}
