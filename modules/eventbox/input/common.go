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

package input

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/types"
	"github.com/erda-project/erda/pkg/dlock"
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
