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

package httputil_test

import (
	"net/http"
	"net/textproto"
	"testing"

	"github.com/erda-project/erda/pkg/http/httputil"
)

func TestHeaderContains(t *testing.T) {
	var h = make(http.Header)
	h.Set("content-type", "application/json; charset=utf-8")
	ok := httputil.HeaderContains(h, httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains(h[textproto.CanonicalMIMEHeaderKey("content-type")], httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains([]httputil.ContentType{
		httputil.ApplicationJson,
		httputil.URLEncodedFormMime,
	}, httputil.ApplicationJson)
	t.Log(ok)

	ok = httputil.HeaderContains("application/json; charset=utf-8", httputil.ApplicationJson)
	t.Log(ok)
}
