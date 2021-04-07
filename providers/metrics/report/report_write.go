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

import "fmt"

type reportWrite struct {
	capacity int
}

func NewReportWrite(capacity int) *reportWrite {
	return &reportWrite{
		capacity: capacity,
	}
}

func (w *reportWrite) Write(data interface{}) error {
	return nil
}
func (w *reportWrite) WriteN(data ...interface{}) (int, error) {
	if w.capacity <= 0 {
		err := fmt.Errorf("buffer max capacity")
		return 0, err
	}
	w.capacity -= len(data)
	return len(data), nil
}
func (w reportWrite) Close() error {
	return nil
}
