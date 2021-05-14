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

package gcdata

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/dlock"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

func GCData() error {
	lock, err := dlock.New("/dice/scheduler/gcdata", func() {})
	if err != nil {
		logrus.Errorf("failed to init gcdata: %v", err)
		return err
	}
	for {
		if err := lock.Lock(context.Background()); err != nil {
			logrus.Errorf("failed to lock /dice/scheduler/gcdata: %v", err)
			lock.Unlock()
			continue
		}
		time.Sleep(24 * time.Hour)
		js, err := jsonstore.New()
		if err != nil {
			logrus.Errorf("failed to new jsonstore: %v", err)
			return err
		}
		gcJobData(js)
		gcServiceData(js)

		lock.Unlock()
	}
}

// gcJobData Clean up pipeline data in etcd
func gcJobData(js jsonstore.JsonStore) error {
	keys, err := js.ListKeys(context.Background(), "/dice/job/")
	if err != nil {
		logrus.Errorf("failed to listkeys: %v", err)
		return err
	}
	for _, key := range keys {
		job := apistructs.Job{}
		if err := js.Get(context.Background(), key, &job); err != nil {
			continue
		}
		if shouldGCJobData(&job) {
			if err := js.Remove(context.Background(), key, nil); err != nil {
				logrus.Errorf("jsonstore remove key: %s, err: %v", key, err)
				break
			}
		}
	}
	// delete /dice/job/<namespace> key-value
	keys, err = js.ListKeys(context.Background(), "/dice/job/")
	if err != nil {
		logrus.Errorf("failed to listkeys: %v", err)
		return err
	}
	for _, key := range keys {
		// /dice/job/<namespace>
		if len(strutil.Split(key, "/", true)) != 3 {
			continue
		}
		subkeys, err := js.ListKeys(context.Background(), key)
		if err != nil {
			logrus.Errorf("failed to listkeys: %v, prefix: %s", err, key)
			continue
		}
		subkeysWithoutPrefixkey := strutil.RemoveSlice(subkeys,
			key,
			strutil.TrimSuffixes(key, "/"),
			strutil.TrimSuffixes(key, "/")+"/")
		if len(subkeysWithoutPrefixkey) != 0 {
			continue
		}
		if err := js.Remove(context.Background(), key, nil); err != nil {
			logrus.Errorf("jsonstore remove key: %s, err: %v", key, err)
			continue
		}
	}
	return nil
}

// gcServiceData Clean up the service data in etcd
func gcServiceData(js jsonstore.JsonStore) {
	// TODO:
}

func shouldGCJobData(job *apistructs.Job) bool {
	if job.CreatedTime == 0 {
		return false
	}
	t := time.Unix(job.CreatedTime, 0)
	before15days := t.Before(time.Now().Add(-20 * 24 * time.Hour))
	return before15days
}
