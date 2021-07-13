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

package openapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/openapi/conf"
)

func NewServer() (*http.Server, error) {
	s, err := NewLoginServer()
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:              conf.ListenAddr(),
		Handler:           s,
		ReadHeaderTimeout: 60 * time.Second, // TODO: test whether will timeout option affect websocket
	}
	return srv, nil
}

func IsHTTPS(req *http.Request) (bool, error) {
	referrer := req.Header.Get("Referer")
	if referrer == "" {
		return false, fmt.Errorf("no Referer header")
	}
	return strings.HasPrefix(referrer, "https:"), nil
}

func replaceProto(isHTTPS bool, v string) string {
	if isHTTPS {
		v = strings.Replace(v, "https", "http", -1)
		v = strings.Replace(v, "http", "https", -1)
	} else {
		v = strings.Replace(v, "https", "http", -1)
	}
	return v
}
