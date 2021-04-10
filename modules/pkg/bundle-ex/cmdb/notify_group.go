// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmdb

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
)

type notifyGroupResp struct {
	apistructs.Header
	Data []*apistructs.NotifyGroup `json:"data"`
}

// QueryNotifyGroup .
func (c *Cmdb) QueryNotifyGroup(groupIDs []string) ([]*apistructs.NotifyGroup, error) {
	var resp *notifyGroupResp
	hresp, err := c.hc.Get(c.url).
		Path("/api/notify-groups/actions/batch-get").
		Header("User-ID", c.operatorID).
		Param("ids", strings.Join(groupIDs, ",")).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !hresp.IsOK() || !resp.Success {
		return nil, fmt.Errorf("status: %d, error: %v", hresp.StatusCode(), resp.Error)
	}
	return resp.Data, nil
}
