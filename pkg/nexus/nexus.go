package nexus

import (
	"time"

	"github.com/erda-project/erda/pkg/httpclient"
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
			httpclient.WithTimeout(time.Second, time.Second*3),
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
