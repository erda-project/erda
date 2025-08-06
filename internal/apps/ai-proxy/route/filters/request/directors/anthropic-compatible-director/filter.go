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
	"net/http/httputil"
	"os"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/anthropic_official"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/anthropic-compatible-director/aws_bedrock"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/common/custom-http-director"
)

var (
	_ filter_define.ProxyRequestRewriter = (*AnthropicCompatibleDirectorRequest)(nil)
)

var RequestRewriterCreator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &AnthropicCompatibleDirectorRequest{
		CustomHTTPDirector: custom_http_director.New(),
		BedrockDirector:    aws_bedrock.NewDirector(),
		AnthropicDirector:  anthropic_official.NewDirector(),
	}
}

func init() {
	filter_define.RegisterFilterCreator("anthropic-compatible-director", RequestRewriterCreator)
}

type AnthropicCompatibleDirectorRequest struct {
	CustomHTTPDirector *custom_http_director.CustomHTTPDirector

	BedrockDirector   *aws_bedrock.BedrockDirector
	AnthropicDirector *anthropic_official.AnthropicDirector
}

func (f *AnthropicCompatibleDirectorRequest) Enable(ctx context.Context) bool {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleAnthropicCompatible))
}

func (f *AnthropicCompatibleDirectorRequest) OnProxyRequest(pr *httputil.ProxyRequest) error {
	if !f.Enable(pr.In.Context()) {
		return nil
	}
	// use custom-http-director to handle the unified http operations
	err := f.CustomHTTPDirector.OnProxyRequest(pr)
	if err != nil {
		return err
	}
	// handle the specific anthropic compatible API style
	apiSegment := api_segment_getter.GetAPISegment(pr.In.Context())
	switch strings.ToLower(string(apiSegment.APIVendor)) {
	case strings.ToLower(string(aws_bedrock.APIVendor)):
		if err := f.BedrockDirector.AwsBedrockDirector(pr, *apiSegment.APIStyleConfig); err != nil {
			return fmt.Errorf("failed to handle aws bedrock director: %v", err)
		}
	case strings.ToLower(string(anthropic_official.APIVendor)):
		if err := f.AnthropicDirector.OfficialDirector(pr, *apiSegment.APIStyleConfig); err != nil {
			return fmt.Errorf("failed to handle anthropic director: %v", err)
		}
	default:
		return fmt.Errorf("unsupported anthropic-compatible api vendor: %s", apiSegment.APIVendor)
	}

	// force set internal pod ip as xff for anthropic region restriction
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		pr.Out.Header.Set("X-Forwarded-For", podIP)
	}

	return nil
}
