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

package util

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/erda-project/erda/pkg/strutil"
)

const Base64EncodedRequestBody = "base64-encoded-request-body"

func HandleRequest(r *http.Request) {
	// base64 decode request body if declared in header
	if strutil.Equal(r.Header.Get(Base64EncodedRequestBody), "true", true) {
		r.Body = io.NopCloser(base64.NewDecoder(base64.StdEncoding, r.Body))
	}
}
