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

package report

import (
	"github.com/erda-project/erda/providers/telemetry/common"
)

type buffer struct {
	count int
	max   int
	data  []*common.Metric
}

func newBuffer(max int) *buffer {
	b := new(buffer)

	b.data = make([]*common.Metric, max+1)
	b.count = 0
	b.max = max

	return b
}

func (b *buffer) IsOverFlow() bool {
	if b.count == b.max {
		return true
	}
	return false
}

func (b *buffer) IsEmpty() bool {
	if b.count == 0 {
		return true
	}
	return false
}

func (b *buffer) Flush() []*common.Metric {
	count := b.count
	b.count = 0
	if count > 0 {
		return b.data[:count]
	}

	return nil
}

func (b *buffer) Add(v *common.Metric) bool {
	if b.count < b.max {
		b.data[b.count] = v
		b.count++
		return true
	}

	return false
}
