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

package cron_test

import (
	"testing"
	"time"

	"github.com/erda-project/erda/internal/pkg/cron"
)

func TestRun(t *testing.T) {
	var count int
	cron.Run()
	stopper, err := cron.Add("*/2 * * * * *", "test", func() bool {
		count++
		t.Logf("[%v][%s] task runs", count, time.Now().Format(time.RFC3339))
		return false
	})
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(time.Second * 5)
	stopper.Stop()
	<-time.After(time.Second * 3)
	if count != 2 && count != 3 {
		t.Fatal("task count error")
	}
}
