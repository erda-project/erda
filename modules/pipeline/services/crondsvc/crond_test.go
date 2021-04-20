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

package crondsvc

import (
	"fmt"
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/cron"
)

// Result:
// 1000  个约 0.03s
// 10000 个约 1.7s
func TestReloadSpeed(t *testing.T) {
	d := cron.New()
	d.Start()
	for i := 0; i < 10; i++ {
		if err := d.AddFunc("*/1 * * * *", func() {
			fmt.Println("hello world")
		}); err != nil {
			panic(err)
		}
	}
	time.Sleep(time.Second * 2)
}
