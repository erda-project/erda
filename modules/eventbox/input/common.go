package input

import (
	"context"
	"time"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/dlock"

	"github.com/sirupsen/logrus"
)

type Handler func(*types.Message) *errors.DispatchError

func OnlyOne(ctx context.Context, lock *dlock.DLock) (func(), error) {
	var isCanceled bool
	var locked bool

	cleanup := func() {
		if isCanceled {
			return
		}
		logrus.Infof("Onlyone: etcdlock: unlock, key: %s", lock.Key())
		if err := lock.Unlock(); err != nil {
			logrus.Errorf("Onlyone: etcdlock unlock err: %v", err)
			return
		}

	}

	go func() {
		time.Sleep(3 * time.Second)
		if !locked && !isCanceled {
			logrus.Warnf("Onlyone: not get lock yet after 3s")
		}
	}()
	if err := lock.Lock(ctx); err != nil {
		if err == context.Canceled {
			isCanceled = true
			logrus.Infof("Onlyone: etcdlock: %v", err)
			return cleanup, nil
		}
		return cleanup, err
	}
	locked = true
	logrus.Infof("Onlyone: etcdlock: lock, key: %s", lock.Key())

	return cleanup, nil
}
