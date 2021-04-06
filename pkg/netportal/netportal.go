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

package netportal

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/customhttp"
)

var bdl *bundle.Bundle

func init() {
	bdl = bundle.New(bundle.WithScheduler())
}

// NewNetportalRequest 在NewRequest基础上增加ClusterName参数，完成netportal代理请求
func NewNetportalRequest(clusterName, method, url string, body io.Reader) (*http.Request, error) {
	info, err := bdl.QueryClusterInfo(clusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	netportalUrl := info.Get(apistructs.NETPORTAL_URL)
	if netportalUrl != "" {
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")
		url = fmt.Sprintf("%s/%s", netportalUrl, url)
	}
	return customhttp.NewRequest(method, url, body)
}
