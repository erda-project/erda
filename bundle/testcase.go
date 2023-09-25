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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) CreateTestCase(req apistructs.TestCaseCreateRequest) (apistructs.AICreateTestCaseResponse, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return apistructs.AICreateTestCaseResponse{}, err
	}
	var resp apistructs.TestCaseCreateResponse
	r, err := b.hc.Post(host).Path("/api/testcases").
		Header(httputil.InternalHeader, "AI").
		Header(httputil.UserHeader, req.IdentityInfo.UserID).
		JSONBody(&req).Do().JSON(&resp)

	if err != nil {
		log.Errorf("CreateTestCase err: %v", err)
		return apistructs.AICreateTestCaseResponse{}, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		log.Errorf("CreateTestCase response is not ok with StatusCode %d", r.StatusCode())
		return apistructs.AICreateTestCaseResponse{}, apierrors.ErrInvoke.InternalError(errors.Errorf("CreateTestCase response is not ok with StatusCode %d", r.StatusCode()))
	}

	return apistructs.AICreateTestCaseResponse{
		TestCaseID: resp.Data,
	}, err
}
