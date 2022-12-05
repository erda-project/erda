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

package match

import (
	"bytes"
	"io"
	"net/http"

	"github.com/erda-project/erda/pkg/encoding/jsonparse"
)

func init() {
	registry("request_body", requestBody{})
}

type requestBody struct {
}

func (f requestBody) get(expr string, r *http.Request) interface{} {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return jsonparse.FilterJson(body, expr, "jsonpath")
}
