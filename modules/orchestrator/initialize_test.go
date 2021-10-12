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

package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func Test_initCron(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	runner := NewMockSharedCronjobRunner(ctrl)
	// call sync addon twice, 1 for first call 2 for loop call
	runner.EXPECT().SyncAddons().Times(2)
	runner.EXPECT().SyncProjects().Times(2)

	initCron(runner, ctx)
}
