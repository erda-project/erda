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
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/manager"
)

const logPrefixContinueBackupQueueUsage = "[queue usage backup]"

// loadQueueManger
func (r *Reconciler) loadQueueManger(ctx context.Context) error {
	// init queue manager
	r.QueueManager = manager.New(ctx, manager.WithDBClient(r.dbClient), manager.WithJSClient(r.js))

	return nil
}

func (r *Reconciler) continueBackupQueueUsage(ctx context.Context) {
	done := make(chan struct{})
	errDone := make(chan error)

	var costTime time.Duration
	for {
		go func() {
			begin := time.Now()
			backup := r.QueueManager.Export()
			end := time.Now()
			costTime = end.Sub(begin)
			queueSnapshot := manager.SnapshotObj{}
			if err := json.Unmarshal(backup, &queueSnapshot); err != nil {
				errDone <- err
			}
			for qID, qMsg := range queueSnapshot.QueueUsageByID {
				if err := r.js.Put(ctx, manager.MakeQueueUsageBackupKey(qID), qMsg); err != nil {
					errDone <- err
					return
				}
			}
			done <- struct{}{}
		}()

		select {
		case <-done:
			logrus.Debugf("%s: sleep 30s for next backup (cost %s this time)", logPrefixContinueBackupQueueUsage, costTime)
			time.Sleep(time.Second * 30)
		case err := <-errDone:
			logrus.Errorf("%s: failed to load, wait 10s for next loading, err: %v", logPrefixContinueBackupQueueUsage, err)
			time.Sleep(time.Second * 10)
		case <-ctx.Done():
			return
		}
	}
}
