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

package metric

import (
	"net/http"
)

type MetricService interface {
	proxy(metric, agg string, request *http.Request) (interface{}, error)
	proxyBlocks(scopeId, id string, request *http.Request) (interface{}, error)
	proxyGroups(scopeId, id string, request *http.Request) (interface{}, error)
}

type TerminusKeyMap struct {
	terminusKey        string `gorm:"column:terminus_key"`
	terminusKeyRuntime string `gorm:"column:terminus_key_runtime"`
}
