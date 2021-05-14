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
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

func testF(text string) error {
	return errors.New(text)
}

func testFF() {
	fmt.Println(time.Now())
}

//func TestDoWithInterval(t *testing.T) {
//	err := DoWithInterval(func() error {
//		testFF()
//		return nil
//	}, 2, time.Second*3)
//	require.Error(t, err)
//}

func TestDo(t *testing.T) {
	var i = 0
	err := DoWithInterval(func() error {
		if i == 0 {
			i++
			return fmt.Errorf("1")
		}
		i++
		return nil
	}, 1, 1*time.Second)
	spew.Dump(err)
}
