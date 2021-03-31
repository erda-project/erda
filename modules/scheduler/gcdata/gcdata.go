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

// gcJobData 清理在 etcd 中的 pipeline 数据
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

// gcServiceData 清理在 etcd 中的 service 数据
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
