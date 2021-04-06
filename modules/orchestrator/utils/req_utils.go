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

package utils

import (
	"math"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
)

func GetPageInfo(r *http.Request) (apistructs.PageInfo, error) {
	page := apistructs.PageInfo{}
	pageNO_ := r.URL.Query().Get("pageNo")
	if len(pageNO_) > 0 {
		pageNO, err := strconv.Atoi(pageNO_)
		if err != nil {
			return page, err
		}
		page.PageNO = pageNO
	}
	pageSize_ := r.URL.Query().Get("pageSize")
	if len(pageSize_) > 0 {
		pageSize, err := strconv.Atoi(pageSize_)
		if err != nil {
			return page, err
		}
		page.PageSize = pageSize
	}
	if page.PageNO <= 0 {
		page.PageNO = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	if page.PageSize > 200 {
		page.PageSize = 200
	}
	return page, nil
}

func Smaller(a, b float64) bool {
	return math.Max(a, b) == b && math.Abs(a-b) > 0.00001
}
