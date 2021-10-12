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

package bundle

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (b *Bundle) JacocoStart(addr string, req *apistructs.JacocoRequest) error {
	hc := httpclient.New(
		httpclient.WithCompleteRedirect(),
		httpclient.WithTimeout(time.Second*3, time.Second*3),
	)
	var jacocoResp apistructs.JacocoResponse
	resp, err := hc.Post(addr).
		Path("/api/jacoco/actions/start").JSONBody(req).
		Do().JSON(&jacocoResp)
	if err != nil {
		return err
	}
	if !resp.IsOK() || !jacocoResp.Success {
		return toAPIError(resp.StatusCode(), jacocoResp.Error)
	}
	return nil
}

func (b *Bundle) JacocoEnd(addr string, req *apistructs.JacocoRequest) error {
	hc := httpclient.New(
		httpclient.WithCompleteRedirect(),
		httpclient.WithTimeout(time.Second*3, time.Second*3),
	)
	var jacocoResp apistructs.JacocoResponse
	resp, err := hc.Post(addr).
		Path("/api/jacoco/actions/end").JSONBody(req).
		Do().JSON(&jacocoResp)
	if err != nil {
		return err
	}
	if !resp.IsOK() || !jacocoResp.Success {
		return toAPIError(resp.StatusCode(), jacocoResp.Error)
	}
	return nil
}
