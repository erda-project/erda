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

package transports

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// RequestFilterGeneratedResponseTransport not actually invoke the backend service, but just return a response from
// agreed ctx key.
type RequestFilterGeneratedResponseTransport struct {
	Inner http.RoundTripper
}

const SchemeForFilterGeneratedResponse = "filter-generated-response"

func (t *RequestFilterGeneratedResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Inner == nil {
		t.Inner = BaseTransport
	}
	// passthrough normal request
	if req.URL.Scheme != SchemeForFilterGeneratedResponse {
		return t.Inner.RoundTrip(req)
	}
	if resp, ok := ctxhelper.GetRequestFilterGeneratedResponse(req.Context()); ok {
		// update request to latest ProxyRequest.Out
		resp.Request = req
		return resp, nil
	}
	return nil, fmt.Errorf("request-filter-generated response not found")
}

func TriggerRequestFilterGeneratedResponse(req *http.Request, resp *http.Response) {
	req.URL.Scheme = SchemeForFilterGeneratedResponse
	ctxhelper.PutRequestFilterGeneratedResponse(req.Context(), resp)
}
