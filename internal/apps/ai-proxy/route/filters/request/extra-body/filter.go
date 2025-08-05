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

package extra_body

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "extra-body"
)

var (
	_ filter_define.ProxyRequestRewriter = (*ExtraBody)(nil)
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type ExtraBody struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &ExtraBody{}
}

func (f *ExtraBody) OnProxyRequest(pr *httputil.ProxyRequest) error {
	// handle json body
	if strings.HasPrefix(pr.Out.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		if err := f.setExtraJSONBody(pr); err != nil {
			return fmt.Errorf("failed to set extra json body, err: %v", err)
		}
		return nil
	}

	// handle other type of body, just pass through
	return nil
}

func (f *ExtraBody) setExtraJSONBody(pr *httputil.ProxyRequest) error {
	// handle extra json body
	bodyCopy, err := body_util.SmartCloneBody(&pr.Out.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}
	var jsonBody map[string]any
	if err := json.NewDecoder(bodyCopy).Decode(&jsonBody); err != nil {
		return fmt.Errorf("failed to decode request body, err: %v", err)
	}
	commonModelMeta := metadata.FromProtobuf(ctxhelper.MustGetModel(pr.Out.Context()).Metadata)
	if err := FulfillExtraJSONBody(&commonModelMeta, ctxhelper.GetIsStream(pr.Out.Context()), jsonBody); err != nil {
		return fmt.Errorf("failed to fulfill extra json body, err: %v", err)
	}
	b, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	if err := body_util.SetBody(pr.Out, b); err != nil {
		return fmt.Errorf("failed to set request body: %w", err)
	}

	return nil
}
