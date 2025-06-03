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

package dashscope_director

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_style"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "dashscope-director"
)

var (
	_ reverseproxy.RequestFilter = (*DashScopeDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type DashScopeDirector struct {
	*reverseproxy.DefaultResponseFilter

	lastCompletionDataLineIndex int
	lastCompletionDataLineText  string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	return &DashScopeDirector{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
	}, nil
}

func (f *DashScopeDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *DashScopeDirector) Enable(ctx context.Context, req *http.Request) bool {
	prov, ok := ctxhelper.GetModelProvider(ctx)
	_, hasAPIConfig := prov.Metadata.Public["api"]
	return ok && strings.EqualFold(prov.Type, string(api_style.APIStyleAliyunDashScope)) && !hasAPIConfig
}
