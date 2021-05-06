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

package helper

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/modules/gittar/webcontext"
)

// HeaderNoCache writing function
func headerNoCache(c *webcontext.Context) {
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	c.Header("cache-Control", "no-cache, max-age=0, must-revalidate")
}

// HeaderCacheForever writing function
func headerCacheForever(c *webcontext.Context) {
	now := time.Now().Unix()
	expires := now + 31536000
	c.Header("Date", fmt.Sprintf("%d", now))
	c.Header("Expires", fmt.Sprintf("%d", expires))
	c.Header("cache-Control", "public, max-age=31536000")
}
