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

package reconciler

import (
	"context"
	"time"

	"github.com/erda-project/erda/pkg/retry"
)

func (r *Reconciler) beforeListen(ctx context.Context) error {
	// init before listen
	if err := retry.DoWithInterval(func() error { return r.loadQueueManger(ctx) }, 3, time.Second*10); err != nil {
		return err
	}
	if err := retry.DoWithInterval(func() error { return r.loadThrottler(ctx) }, 3, time.Second*10); err != nil {
		return err
	}
	return nil
}
