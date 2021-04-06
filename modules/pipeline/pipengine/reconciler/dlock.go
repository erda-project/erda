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
