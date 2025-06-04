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

package openai_compatible_director

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	custom_http_director "github.com/erda-project/erda/internal/apps/ai-proxy/filters/directors/custom-http-director"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_segment_getter"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "openai-compatible-director"
)

var (
	_ reverseproxy.RequestFilter = (*OpenaiCompatibleDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type OpenaiCompatibleDirector struct {
	*custom_http_director.CustomHTTPDirector
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &OpenaiCompatibleDirector{CustomHTTPDirector: custom_http_director.New()}, nil
}

func (f *OpenaiCompatibleDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *OpenaiCompatibleDirector) Enable(ctx context.Context, _ *http.Request) bool {
	apiSegment := api_segment_getter.GetAPISegment(ctx)
	return apiSegment != nil &&
		strings.EqualFold(string(apiSegment.APIStyle), string(api_style.APIStyleOpenAICompatible))
}
