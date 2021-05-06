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

package marathon

import (
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/httpclient"
)

func TestMarathon_SuspendApp(t *testing.T) {
	m := &Marathon{
		name:   "MARATHONFORTERMINUSTEST",
		addr:   "dcos.test.terminus.io/service/marathon",
		client: httpclient.New().BasicAuth("admin", "Terminus1234"),
	}

	ch := make(chan string, 10)
	for i := 0; i < 1; i++ {
		ch <- "/runtimes/v1/addon-redis/d140959eb05b4960846e7b0fb3cc42a6/redis-master-d140959eb05b4960846e7b0fb3cc42a6"
		time.Sleep(1 * time.Second)
	}
	go m.SuspendApp(ch)

	time.Sleep(1 * time.Second)
	close(ch)
}
