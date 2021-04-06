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
	"io"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// CollectLogs 收集日志
func (b *Bundle) CollectLogs(source string, body io.Reader) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc
	resp, err := hc.Post(host).Path(strutil.Concat("/collect/logs/", source)).
		Header("Internal-Client", "bundle").
		RawBody(body).
		Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("failed to call monitor status %d", resp.StatusCode()))
	}
	return nil
}

// CollectMetrics 收集指标
func (b *Bundle) CollectMetrics(metrics *apistructs.Metrics) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc
	resp, err := hc.Post(host).Path("/collect/metrics").
		Header("Internal-Client", "bundle").JSONBody(&metrics).Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("failed to call monitor status %d", resp.StatusCode()))
	}
	return nil
}
