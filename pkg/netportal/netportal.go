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
	"os"
	"strings"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/customhttp"
)

var bdl *bundle.Bundle

func init() {
	bdl = bundle.New(bundle.WithScheduler())
}

// NewNetportalRequest 在NewRequest基础上增加ClusterName参数，完成netportal代理请求
func NewNetportalRequest(clusterName, method, url string, body io.Reader) (*http.Request, error) {
	currentClusterName := os.Getenv("DICE_CLUSTER_NAME")
	if currentClusterName != clusterName {
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")
		url = fmt.Sprintf("inet://%s/%s", clusterName, url)
	}
	return customhttp.NewRequest(method, url, body)
}
