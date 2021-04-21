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

package logic

import (
	"context"

	"github.com/erda-project/erda/apistructs"
)

// getNetportalURL return netportalURL or empty string.
func getNetportalURL(ctx context.Context) string {
	if u := ctx.Value(apistructs.NETPORTAL_URL); u != nil {
		if s, ok := u.(string); ok {
			return s
		}
	}
	return ""
}
