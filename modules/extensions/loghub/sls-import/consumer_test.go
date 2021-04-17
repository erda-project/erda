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
