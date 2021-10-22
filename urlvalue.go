//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"net/url"
	"strings"
)

func main() {
	v := make(url.Values)
	namespaces := []string{"default", "n1", "n2"}
	v.Add("erda-hongkong", strings.Join(namespaces, ","))
	v.Add("erda-cloud", strings.Join(namespaces, ","))
	fmt.Println(v.Encode())

	s := v.Encode()
	values, err := url.ParseQuery(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(values)
}
