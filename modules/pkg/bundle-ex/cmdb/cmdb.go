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

package cmdb

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

// Cmdb .
type Cmdb struct {
	url        string
	operatorID string
	hc         *httpclient.HTTPClient
}

// Option .
type Option func(cmdb *Cmdb)

// New .
func New(options ...Option) *Cmdb {
	addr := strings.TrimRight(discover.CoreServices(), "/")
	opid := os.Getenv("DICE_OPERATOR_ID")
	if len(opid) <= 0 {
		opid = "1100"
	}
	cmdb := &Cmdb{
		url:        fmt.Sprintf("http://%s", addr),
		operatorID: opid,
	}
	for _, op := range options {
		op(cmdb)
	}
	if cmdb.hc == nil {
		cmdb.hc = httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	}
	return cmdb
}

// WithURL .
func WithURL(url string) Option {
	return func(e *Cmdb) {
		e.url = strings.TrimRight(url, "/")
	}
}

// WithOperatorID .
func WithOperatorID(operatorID string) Option {
	return func(e *Cmdb) {
		e.operatorID = operatorID
	}
}

// WithHTTPClient .
func WithHTTPClient(hc *httpclient.HTTPClient) Option {
	return func(e *Cmdb) {
		e.hc = hc
	}
}
