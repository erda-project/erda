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

package plugins

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
)

// Interface .
type Interface interface {
	Validate(c *pb.Checker) error
	New(c *pb.Checker) (Handler, error)
}

// Checker .
type Handler interface {
	Do(pctx Context) error
}

// Context .
type Context interface {
	context.Context
	Report(m ...*Metric) error
}

// Metric .
type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
	Tags      map[string]string      `json:"tags"`
}
