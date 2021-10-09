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
	"github.com/erda-project/erda/apistructs"
)

func getJacocoAddr() string {
	return "http://30.43.49.4:8801"
}

func (b *Bundle) JacocoStart(req *apistructs.JacocoRequest) error {
	host := getJacocoAddr()
	hc := b.hc
	var jacocoResp apistructs.JacocoResponse
	resp, err := hc.Post(host).
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

func (b *Bundle) JacocoEnd(req *apistructs.JacocoRequest) error {
	host := getJacocoAddr()
	hc := b.hc
	var jacocoResp apistructs.JacocoResponse
	resp, err := hc.Post(host).
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
