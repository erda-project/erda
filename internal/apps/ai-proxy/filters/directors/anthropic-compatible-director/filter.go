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
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/anthropic_official"
	"github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/anthropic-compatible-director/aws_bedrock"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/custom-http-director"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "anthropic-compatible-director"
)

var (
	_ reverseproxy.RequestFilter = (*AnthropicCompatibleDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type AnthropicCompatibleDirector struct {
	CustomHTTPDirector *custom_http_director.CustomHTTPDirector

	*reverseproxy.DefaultResponseFilter
	BedrockDirector   *aws_bedrock.BedrockDirector
	AnthropicDirector *anthropic_official.AnthropicDirector
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &AnthropicCompatibleDirector{
		CustomHTTPDirector:    custom_http_director.New(),
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
		BedrockDirector:       aws_bedrock.NewDirector(),
		AnthropicDirector:     anthropic_official.NewDirector(),
	}, nil
}

func (f *AnthropicCompatibleDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *AnthropicCompatibleDirector) Enable(ctx context.Context, _ *http.Request) bool {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleAnthropicCompatible))
}

func (f *AnthropicCompatibleDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// use custom-http-director to handle the unified http operations
	signal, err = f.CustomHTTPDirector.OnRequest(ctx, w, infor)
	if signal == reverseproxy.Intercept || err != nil {
		return reverseproxy.Intercept, err
	}
	// handle the specific anthropic compatible API style
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		if err := f.BedrockDirector.AwsBedrockDirector(ctx, infor, *apiSegment.APIStyleConfig); err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to handle aws bedrock director: %v", err)
		}
	case strings.ToLower(string(anthropic_official.APIVendor)):
		if err := f.AnthropicDirector.OfficialDirector(ctx, infor, *apiSegment.APIStyleConfig); err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to handle anthropic director: %v", err)
		}
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}
	return reverseproxy.Continue, nil
}

func (f *AnthropicCompatibleDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		return f.BedrockDirector.OnResponseChunk(ctx, infor, w, chunk)
	case strings.ToLower(string(anthropic_official.APIVendor)):
		return f.AnthropicDirector.OnResponseChunk(ctx, infor, w, chunk)
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}
}

func (f *AnthropicCompatibleDirector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		return f.BedrockDirector.OnResponseEOF(ctx, infor, w, chunk)
	case strings.ToLower(string(anthropic_official.APIVendor)):
		return f.AnthropicDirector.OnResponseEOF(ctx, infor, w, chunk)
	default:
		return fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}
}
