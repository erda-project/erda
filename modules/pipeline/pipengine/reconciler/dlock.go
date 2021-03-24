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
