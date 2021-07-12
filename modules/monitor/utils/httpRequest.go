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

package utils

import (
	"context"
	"net/http"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
)

func GetHttpRequest(ctx context.Context) *http.Request {
	return transhttp.ContextRequest(ctx)
}

func NewContextWithHeader(ctx context.Context) context.Context {
	httpRequest := GetHttpRequest(ctx)
	header := transport.Header{}
	for k := range httpRequest.Header {
		header.Set(k, httpRequest.Header.Get(k))
	}
	return transport.WithHeader(context.Background(), header)
}
