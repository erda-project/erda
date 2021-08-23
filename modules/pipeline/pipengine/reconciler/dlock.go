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
	"strconv"

	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/strutil"
)

func makeReconcilerDLockKey(pipelineID uint64) (string, error) {
	return strutil.Concat(etcdReconcilerDLockPrefix, strconv.FormatUint(pipelineID, 10)), nil
}

// lockPipeline for push
func lockPipeline(ctx context.Context, pipelineID uint64, dLockLostFunc func()) (*dlock.DLock, error) {
	lockKey, err := makeReconcilerDLockKey(pipelineID)
	if err != nil {
		return nil, err
	}
	lock, err := dlock.New(lockKey, dLockLostFunc, dlock.WithTTL(30))
	if err != nil {
		return nil, err
	}
	if err := lock.Lock(ctx); err != nil {
		return nil, err
	}
	return lock, nil
}
