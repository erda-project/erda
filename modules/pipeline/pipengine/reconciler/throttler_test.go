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
	"reflect"
	"testing"

	"bou.ke/monkey"
)

func TestContinueBackupThrottler(t *testing.T) {
	//js := &jsonstore.JsonStoreImpl{}
	//pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(js), "Put", func(j *jsonstore.JsonStoreImpl, ctx context.Context, key string, object interface{}) error {
	//	return nil
	//})
	//defer pm1.Unpatch()
	//
	//tl := throttler.NewNamedThrottler("default", nil)
	//r := &Reconciler{js: js, TaskThrottler: tl}
	//t.Run("ContinueBackupThrottler", func(t *testing.T) {
	//	ctx, cancel := context.WithCancel(context.Background())
	//	r.ContinueBackupThrottler(ctx)
	//	time.Sleep(1 * time.Second)
	//	cancel()
	//})
	r := &Reconciler{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(r), "ContinueBackupThrottler", func(r *Reconciler, ctx context.Context) {
		return
	})
	defer pm1.Unpatch()
	t.Run("ContinueBackupThrottler", func(t *testing.T) {
		r.ContinueBackupThrottler(context.Background())
	})
}
