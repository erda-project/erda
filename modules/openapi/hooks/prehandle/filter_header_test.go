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

package prehandle

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/httputil"
)

func TestFilterHeader(t *testing.T) {
	r := http.Request{
		Header: make(http.Header),
	}

	// add User-ID header
	r.Header.Add(httputil.UserHeader, "1")
	r.Header.Add(httputil.UserHeader, "2")
	assert.True(t, 2 == len(r.Header.Values(httputil.UserHeader)))
	r.Header.Del(httputil.UserHeader)
	assert.True(t, 0 == len(r.Header.Values(httputil.UserHeader)))
	r.Header.Del(httputil.UserHeader)
	assert.True(t, 0 == len(r.Header.Values(httputil.UserHeader)))

	// add Internal-Client header
	r.Header.Add(httputil.InternalHeader, "bundle")
	assert.True(t, 1 == len(r.Header.Values(httputil.InternalHeader)))
	r.Header.Add(httputil.InternalHeader, "true")
	assert.True(t, 2 == len(r.Header.Values(httputil.InternalHeader)))

	// add Client-ID/Client-Name header
	r.Header.Add(httputil.ClientIDHeader, "bundle")
	r.Header.Add(httputil.ClientNameHeader, "bundle")

	// add a valid custom header
	r.Header.Add("Pipeline-ID", "1")

	// filter
	FilterHeader(context.Background(), nil, &r)

	assert.True(t, 1 == len(r.Header))
}
