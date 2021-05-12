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
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/pkg/httpclient"
)

// PushLog 推日志
func (b *Bundle) GetAkSkByAk(ak string) (model.AkSk, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return model.AkSk{}, err
	}
	hc := b.hc

	var obj model.AkSk
	resp, err := hc.Get(host, httpclient.RetryErrResp).
		Path("/api/aksks/"+ak).
		Header("Content-Type", "application/json").
		Do().JSON(&obj)
	if err != nil || !resp.IsOK() {
		return model.AkSk{}, apierrors.ErrInvoke.InternalError(err)
	}

	return obj, nil
}
