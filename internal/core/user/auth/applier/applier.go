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

package applier

import (
	"net/http"

	"github.com/erda-project/erda/pkg/http/httpclient"
)

type CookieTokenAuth struct {
	Header http.Header
}

func (c *CookieTokenAuth) Apply(req *httpclient.Request) {
	req = req.Headers(c.Header)
}

type BearerTokenAuth struct {
	Token string
}

func (b *BearerTokenAuth) Apply(req *httpclient.Request) {
	req = req.Header("Authorization", "Bearer "+b.Token)
}

type QueryTokenAuth struct {
	Param string
	Token string
}

func (q *QueryTokenAuth) Apply(req *httpclient.Request) {
	req.Param(q.Param, q.Token)
}
