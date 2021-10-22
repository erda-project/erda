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

package tasks_test

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/cmp/tasks"
)

func TestExitError_Error(t *testing.T) {
	var e = tasks.ExitError{Msg: "something wrong"}
	t.Log(e.Error())
}

func TestTicker_Close(t *testing.T) {
	var times = 0
	ticker := tasks.New(time.Second*2, func() error {
		times++
		if times > 3 {
			return errors.New("normal error")
		}
		if times > 5 {
			return tasks.ExitError{Msg: "time over"}
		}
		t.Log("times:", times)
		return nil
	})
	ticker.Run()
}
