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

package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/erda-project/erda/internal/pkg/reverseproxy"
)

func main() {
	var upstreamHost = "127.0.0.1:8080"
	var listenAndServe = ":8082"
	http.DefaultServeMux.Handle("/", &reverseproxy.ReverseProxy{
		Director: func(r *http.Request) {
			r.Host = upstreamHost
			r.URL.Host = upstreamHost
			r.URL.Scheme = "http"
		},
		Transport:      nil,
		FlushInterval:  time.Millisecond * 100,
		ErrorLog:       nil,
		BufferPool:     nil,
		ModifyResponse: nil,
		ErrorHandler:   nil,
		Filters: []any{
			&filterTrimSpace{},
			&filterAppendSep{},
		},
		FilterCxt: context.Background(),
	})
	log.Printf("ListenAndServe %s, reverse proxy to %s\n", listenAndServe, upstreamHost)
	if err := http.ListenAndServe(":8082", http.DefaultServeMux); err != nil {
		panic(err)
	}
}

var (
	_ reverseproxy.ResponseFilter = (*filterTrimSpace)(nil)
	_ reverseproxy.ResponseFilter = (*filterAppendSep)(nil)
)

type filterTrimSpace struct {
}

func (f *filterTrimSpace) OnResponseChunk(ctx context.Context, response *http.Response, writer io.Writer, in []byte) (next bool, err error) {
	log.Printf("*filterTrimSpace.OnResponseChunk called, in data: %s\n", string(in))
	for _, c := range in {
		if c == ' ' {
			continue
		}
		_, _ = writer.Write([]byte{c})
	}
	return true, nil
}

func (f *filterTrimSpace) OnResponseEOF(ctx context.Context, response *http.Response, writer io.Writer, in []byte) error {
	log.Printf("*filterTrimSpace.OnResponseEOF called, in data: %s\n", string(in))
	_, _ = f.OnResponseChunk(ctx, response, writer, in)
	return nil
}

type filterAppendSep struct {
	data []byte
}

func (f *filterAppendSep) OnResponseChunk(ctx context.Context, response *http.Response, writer io.Writer, in []byte) (next bool, err error) {
	log.Printf("*filterAppendSep.OnResponseChunk called, in data: %s\n", string(in))
	//in = append(in, ';', '\n')
	//f.data = append(f.data, in...)
	_, _ = writer.Write(in)
	return true, nil
}

func (f *filterAppendSep) OnResponseEOF(ctx context.Context, response *http.Response, writer io.Writer, in []byte) error {
	log.Printf("*filterAppendSep.OnResponseEOF called, in data: %s\n", string(in))
	_, _ = f.OnResponseChunk(ctx, response, writer, in)
	log.Printf("filterAppendSep final data: %s\n", string(f.data))
	return nil
}
