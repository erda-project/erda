// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package endpoints_test

import (
	"net/url"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/endpoints"
)

func TestParseApplicationsResourceQuery(t *testing.T) {
	var (
		query  = new(apistructs.ApplicationsResourceQuery)
		values url.Values
	)
	endpoints.ParseApplicationsResourceQuery(query, values)

	values = make(url.Values)
	values.Add("applicationID", "1")
	values.Add("applicationID", "2")
	values.Add("ownerID", "1")
	values.Add("ownerID", "2")
	values.Add("orderBy", "cpuRequest,asc")
	values.Add("orderBy", "memRequest,desc")
	endpoints.ParseApplicationsResourceQuery(query, values)

	values.Add("pageNo", "10")
	values.Add("pageSize", "2")
	endpoints.ParseApplicationsResourceQuery(query, values)

	if query.PageNo != 10 {
		t.Errorf("parse pageNo error, actual: %v", query.PageNo)
	}
	if query.PageSize != 2 {
		t.Errorf("parse pageSize error, actual: %v", query.PageSize)
	}

	t.Logf("%+v", query)
}
