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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
)

// PushLog 推日志
func (b *Bundle) PushLog(req *apistructs.LogPushRequest) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc

	resp, err := hc.Post(host, httpclient.RetryErrResp).
		Path("/collect/logs/job").
		JSONBody(req.Lines).
		Header("Content-Type", "application/json").
		Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to pushLog, error is %d", resp.StatusCode()))
	}
	return nil
}
