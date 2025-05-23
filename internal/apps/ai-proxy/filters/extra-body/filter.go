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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "extra-body"
)

var (
	_ reverseproxy.RequestFilter  = (*ExtraBody)(nil)
	_ reverseproxy.ResponseFilter = (*ExtraBody)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type ExtraBody struct {
	*reverseproxy.DefaultResponseFilter
}

func (f *ExtraBody) Enable(ctx context.Context, req *http.Request) bool {
	return true
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &ExtraBody{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

func (f *ExtraBody) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	// handle json body
	if strings.HasPrefix(infor.Header().Get(httputil.HeaderKeyContentType), string(httputil.ApplicationJson)) {
		if err := f.setExtraJSONBody(ctx, infor); err != nil {
			return reverseproxy.Intercept, fmt.Errorf("failed to set extra json body, err: %v", err)
		}
		return reverseproxy.Continue, nil
	}

	// handle other type of body, just pass through
	return reverseproxy.Continue, nil
}

func (f *ExtraBody) setExtraJSONBody(ctx context.Context, infor reverseproxy.HttpInfor) error {
	// handle extra json body
	var jsonBody map[string]any
	if err := json.NewDecoder(infor.Body()).Decode(&jsonBody); err != nil {
		return fmt.Errorf("failed to decode request body, err: %v", err)
	}
	commonModelMeta := metadata.FromProtobuf(ctxhelper.MustGetModel(ctx).Metadata)
	if err := FulfillExtraJSONBody(&commonModelMeta, ctxhelper.GetIsStream(ctx), jsonBody); err != nil {
		return fmt.Errorf("failed to fulfill extra json body, err: %v", err)
	}
	b, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body, err: %v", err)
	}
	infor.SetBody2(b)

	return nil
}
