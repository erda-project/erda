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

package v2

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

func ReadRequestBody(req *http.Request) ([]byte, error) {
	defer func() {
		req.Body.Close()
	}()

	reader, err := getBodyReader(req)
	if err != nil {
		return []byte{}, err
	}

	reader, err = getCustomEncodingReader(req, reader)
	if err != nil {
		return []byte{}, err
	}

	return io.ReadAll(reader)
}

func getCustomEncodingReader(req *http.Request, reader io.Reader) (io.Reader, error) {
	ccEncoding := req.Header.Get("Custom-Content-Encoding")
	if ccEncoding == "" {
		return reader, nil
	} else if ccEncoding == "base64" {
		return base64.NewDecoder(base64.StdEncoding, reader), nil
	}
	return nil, errors.New("unsupported custom-content-encoding")
}

func getBodyReader(req *http.Request) (io.Reader, error) {
	contentEncoding := req.Header.Get("Content-Encoding")
	if contentEncoding == "gzip" {
		gzipReader, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		return gzipReader, nil
	}
	return req.Body, nil
}
