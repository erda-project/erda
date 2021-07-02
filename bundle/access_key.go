// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package bundle

import (
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
