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

package addon

import (
	"fmt"
	"testing"
	"time"
)

func TestRetention(t *testing.T) {
	//d, err := time.Parse("7d2h1m10s")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(d.Seconds())

	fmt.Println(time.Duration(1000000000 * time.Second).String())
}
