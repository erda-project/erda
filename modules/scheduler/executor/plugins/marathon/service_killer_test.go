// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package marathon

import (
	"testing"
	"time"

	"github.com/erda-project/erda/pkg/http/httpclient"
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
