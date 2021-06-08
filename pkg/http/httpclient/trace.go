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

package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Tracer interface {
	TraceRequest(*http.Request)
	// if read response body, you need to set it back
	TraceResponse(*http.Response)
}

type DefaultTracer struct {
	w io.Writer
}

func NewDefaultTracer(w io.Writer) *DefaultTracer {
	return &DefaultTracer{w}
}

func (t *DefaultTracer) TraceRequest(req *http.Request) {
	s := fmt.Sprintf("RequestURL: %s\n", req.URL.String())
	io.WriteString(t.w, s)
}

func (t *DefaultTracer) TraceResponse(r *http.Response) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		io.WriteString(t.w, fmt.Sprintf("TraceResponse: read response body fail: %v", err))
		return
	}
	io.WriteString(t.w, fmt.Sprintf("ResponseBody: %s\n", string(body)))
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
}
