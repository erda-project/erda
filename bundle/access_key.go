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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

func (b *Bundle) GetAccessKeyByAccessKeyID(ak string) (model.AccessKey, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return model.AccessKey{}, err
	}
	hc := b.hc

	var obj AkSkResponse
	resp, err := hc.Get(host, httpclient.RetryErrResp).
		Path("/api/credential/access-keys/"+ak).
		Header("Content-Type", "application/json").
		Do().JSON(&obj)
	if err != nil || !resp.IsOK() {
		return model.AccessKey{}, apierrors.ErrInvoke.NotFound()
	}

	return obj.Data, nil
}

type AkSkResponse struct {
	Data model.AccessKey `json:"data"`
}

func (b *Bundle) ListAccessKeys(req apistructs.AccessKeyListQueryRequest) ([]model.AccessKey, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var obj AccessKeysListResponse
	r := hc.Get(host, httpclient.RetryErrResp).Path("/api/credential/access-keys").Header("Content-Type", "application/json")
	if req.Subject != "" {
		r = r.Param("subject", req.Subject)
	}
	if req.SubjectType != "" {
		r = r.Param("subjectType", req.SubjectType)
	}
	if req.Status != "" {
		r = r.Param("status", req.Status)
	}
	if req.IsSystem != nil {
		r = r.Param("isSystem", strconv.FormatBool(*req.IsSystem))
	}

	resp, err := r.Do().JSON(&obj)
	if err != nil || !resp.IsOK() {
		return nil, apierrors.ErrInvoke.NotFound()
	}

	return obj.Data, nil
}

type AccessKeysListResponse struct {
	Data []model.AccessKey `json:"data"`
}
