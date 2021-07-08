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
	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"net/http"
)

func GetHttpRequest(ctx context.Context) *http.Request {
	return transhttp.ContextRequest(ctx)
}

func NewContextWithHeader(ctx context.Context) context.Context {
	httpRequest := GetHttpRequest(ctx)
	header := transport.Header{}
	for k, _ := range httpRequest.Header {
		header.Set(k, httpRequest.Header.Get(k))
	}
	return transport.WithHeader(context.Background(), header)
}

//func NewContextWithMetadata(ctx context.Context) context.Context {
//	httpRequest := GetHttpRequest(ctx)
//	header := make(map[string]string)
//	for k,_ := range httpRequest.Header {
//		header.Set(k,httpRequest.Header.Get(k))
//	}
//	md := metadata.New(header)
//	context := metadata.NewOutgoingContext(context.Background(),md)
//	return context
//}

func NewContextWithLang(ctx context.Context) context.Context {
	httpRequest := GetHttpRequest(ctx)
	lang := httpRequest.Header.Get("Lang")
	if len(lang) <= 0 {
		lang = httpRequest.Header.Get("Accept-Language")
	}
	header := transport.Header{}
	header.Set("Lang", lang)
	return transport.WithHeader(context.Background(), header)
}
