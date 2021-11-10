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

package httpclient

import (
	"net/url"
	"testing"
)

func TestSetParams(t *testing.T) {
	var (
		r Request
		p = make(url.Values)
	)
	p.Add("withQuota", "true")
	SetParams(p)(&r)
	if v := r.params.Get("withQuota"); v != "true" {
		t.Fatal("error")
	}
}
