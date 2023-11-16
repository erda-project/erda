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
	"fmt"
	"net/url"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) CreateTestSet(req apistructs.TestSetCreateRequest) (*apistructs.TestSet, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return &apistructs.TestSet{}, err
	}

	var data apistructs.TestSetCreateResponse
	r, err := b.hc.Post(host).Path("/api/testsets").
		Header(httputil.InternalHeader, "AI").
		Header(httputil.UserHeader, req.IdentityInfo.UserID).
		JSONBody(&req).Do().JSON(&data)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "CreateTestSet failed"})
	}
	return data.Data, nil
}

func (b *Bundle) GetTestSets(req apistructs.TestSetListRequest) ([]apistructs.TestSet, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("recycled", "false")
	params.Set("parentID", fmt.Sprintf("%d", *req.ParentID))
	params.Set("projectID", fmt.Sprintf("%d", *req.ProjectID))

	var data apistructs.TestSetListResponse
	r, err := b.hc.Get(host).Path("/api/testsets").
		Params(params).
		Header(httputil.InternalHeader, "AI").
		JSONBody(&req).Do().JSON(&data)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "GetTestSets failed"})
	}

	return data.Data, nil
}

// GET {Path: "/api/testsets/{testSetID}", Method: http.MethodGet, Handler: e.GetTestSet},
func (b *Bundle) GetTestSetById(req apistructs.TestSetGetRequest) (*apistructs.TestSetWithAncestors, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return nil, err
	}

	var data apistructs.TestSetGetResponse
	path := fmt.Sprintf("/api/testsets/%d", req.ID)
	r, err := b.hc.Get(host).Path(path).
		Header(httputil.InternalHeader, "AI").
		Header(httputil.UserHeader, req.IdentityInfo.UserID).
		JSONBody(&req).Do().JSON(&data)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "GetTestSetById failed"})
	}

	return data.Data, nil
}
