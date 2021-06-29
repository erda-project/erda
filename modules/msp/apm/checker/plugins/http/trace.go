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

package http

import (
	"math/rand"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func (p *provider) shouldSample() bool {
	if p.Cfg.Trace.Enable && p.Cfg.Trace.Rate > 0 {
		if random.Intn(100) < p.Cfg.Trace.Rate {
			return true
		}
	}
	return false
}

func (p *provider) setupTrace(req *http.Request, tags map[string]string, fields map[string]interface{}) {
	requestID := uuid.NewV4().String()
	tags["request_id"] = requestID
	req.Header.Set("terminus-request-id", requestID)
	req.Header.Set("terminus-request-id", "true")
}
