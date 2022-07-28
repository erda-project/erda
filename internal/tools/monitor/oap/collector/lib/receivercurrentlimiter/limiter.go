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

package receivercurrentlimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
)

var (
	// Limit the max number of request which is handling by receiver.
	maxCurrentRequest = lib.AvailableCPUs() * 4
	// Limit the duration of request which wait for handle.
	maxQueueDuration = 30 * time.Second
)

var ch chan struct{}

func Init() {
	ch = make(chan struct{}, maxCurrentRequest)
}

func Do(fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueueDuration)
	defer cancel()
	select {
	case ch <- struct{}{}:
		err := fn()
		<-ch
		return err
	case <-ctx.Done():
		return fmt.Errorf("unable to handle more request, pls retry")
	}
}
