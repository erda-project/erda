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

package slsimport

import (
	"fmt"
)

func Example_http_status() {
	status, typ, err := parseHTTPStatus("201")
	fmt.Println(status, typ, err)

	status, typ, err = parseHTTPStatus("404")
	fmt.Println(status, typ, err)

	status, typ, err = parseHTTPStatus("599")
	fmt.Println(status, typ, err)

	status, typ, err = parseHTTPStatus("1990")
	fmt.Println(status, typ, err)

	// Output:
	// 201 2XX <nil>
	// 404 4XX <nil>
	// 599 5XX <nil>
	// 1990 19XX <nil>
}
