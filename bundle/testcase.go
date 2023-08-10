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
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) CreateTestCase(req apistructs.TestCaseCreateRequest) ([]byte, error) {
	resp, err := b.hc.Post("localhost:9095").Path("/api/testcases").
		Header(httputil.InternalHeader, "AI").
		Header(httputil.UserHeader, req.IdentityInfo.UserID).
		JSONBody(&req).Do().RAW()
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	log.Printf("CreateTestCase response data: %s, err: %v", string(data), err)
	return data, err
}
