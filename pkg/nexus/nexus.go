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

package nexus

import (
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type Nexus struct {
	Server

	hc             *httpclient.HTTPClient
	blobNetdataDir string
}

type Server struct {
	Addr     string
	Username string
	Password string
}

type Option func(*Nexus)

func New(server Server, ops ...Option) *Nexus {
	n := Nexus{
		Server: server,
	}
	for _, op := range ops {
		op(&n)
	}
	if n.hc == nil {
		n.hc = httpclient.New(
		// 3 seconds is too short, temporarily change to the default 60 seconds
		// Decouple the connection between publisher and nexsus in the future
		// httpclient.WithTimeout(time.Second, time.Second*3),
		)
	}
	if n.blobNetdataDir == "" {
		n.blobNetdataDir = DefaultBlobNetdataDir
	}
	return &n
}

func WithHttpClient(hc *httpclient.HTTPClient) Option {
	return func(n *Nexus) {
		n.hc = hc
	}
}

func WithBlobNetdataDir(dir string) Option {
	return func(n *Nexus) {
		n.blobNetdataDir = dir
	}
}
