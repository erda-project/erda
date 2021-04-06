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

package retry

import (
	"time"

	"github.com/hashicorp/go-multierror"
)

func Do(fn func() error, n int) (err error) {
	return DoWithInterval(fn, n, 0)
}

func DoWithInterval(fn func() error, n int, interval time.Duration) error {
	var me *multierror.Error
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		err := fn()
		if err == nil {
			me = nil
			break
		}
		me = multierror.Append(me, err)
		time.Sleep(interval)
	}
	if me != nil {
		return me.ErrorOrNil()
	}
	return nil
}
