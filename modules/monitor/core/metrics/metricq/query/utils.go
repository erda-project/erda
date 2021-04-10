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

package query

import (
	"fmt"
	"strings"

	"github.com/recallsong/go-utils/encoding/jsonx"
)

// ElasticSearchCURL .
func ElasticSearchCURL(url string, indices []string, source interface{}) string {
	body := jsonx.MarshalAndIntend(source)
	body = strings.Replace(body, "'", "'\\''", -1)
	return fmt.Sprintf(`
curl -X GET \
'%s/%s/_search' \
-H 'Content-Type: application/json' \
-H 'cache-control: no-cache' \
-d '%s'
`, url, strings.Join(indices, ","), body)
}
