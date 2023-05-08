// Copyright (c) 2023 Terminus, Inc.
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

package reverseproxy

import (
	"bytes"
	"io"
	"net/http"
)

var (
	_ http.RoundTripper = (*DoNothingTransport)(nil)
)

type DoNothingTransport struct {
	Response *http.Response
}

func (d *DoNothingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.Response == nil {
		d.Response = &http.Response{
			Status:           "",
			StatusCode:       0,
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           make(http.Header),
			Body:             io.NopCloser(bytes.NewReader(nil)),
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          req,
			TLS:              nil,
		}
	}
	return d.Response, nil
}
