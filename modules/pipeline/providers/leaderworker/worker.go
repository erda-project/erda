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

package leaderworker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) listWorkersByType(ctx context.Context, typ worker.Type) ([]worker.Worker, error) {
	getResp, err := p.EtcdClient.Get(ctx, p.makeEtcdWorkerKeyPrefix(typ), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var workers []worker.Worker
	for _, kv := range getResp.Kvs {
		w, err := worker.NewFromBytes(kv.Value)
		if err != nil {
			return nil, err
		}
		w.SetType(typ)
		workers = append(workers, w)
	}
	return workers, nil
}

func (p *provider) getWorkerTaskLogicIDFromIncomingKey(workerID worker.ID, key string) worker.LogicTaskID {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	return worker.LogicTaskID(strutil.TrimPrefixes(key, prefix))
}

// see: makeEtcdWorkerTaskDispatchKey
func (p *provider) getWorkerIDFromIncomingKey(key string) worker.ID {
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	if !strutil.HasPrefixes(key, prefix) {
		return ""
	}
	workerIDAndSuffix := strutil.TrimPrefixes(key, prefix)
	workerIDAndLogicTaskID := strutil.Split(workerIDAndSuffix, "/task/")
	if len(workerIDAndLogicTaskID) != 2 {
		return ""
	}
	return worker.ID(workerIDAndLogicTaskID[0])
}

func (p *provider) getWorkerLastHeartbeatUnixTime(ctx context.Context, workerID worker.ID) (lastHeartbeatUnitTime int64, err error, needRetry bool) {
	getResp, err := p.EtcdClient.Get(ctx, p.makeEtcdWorkerHeartbeatKey(workerID))
	if err != nil {
		return 0, err, true
	}
	if len(getResp.Kvs) == 0 {
		return 0, fmt.Errorf("no heartbeat info"), false
	}
	lastHeartbeatTime, err := strconv.ParseInt(string(getResp.Kvs[0].Value), 10, 64)
	if err != nil {
		return 0, err, false
	}
	return lastHeartbeatTime, nil, false
}

func (p *provider) getWorkerFromEtcd(ctx context.Context, workerID worker.ID, typ worker.Type) (worker.Worker, error) {
	key := p.makeEtcdWorkerKey(workerID, typ)
	getResp, err := p.EtcdClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if getResp == nil || getResp.Count == 0 || len(getResp.Kvs) == 0 {
		return nil, worker.ErrNotFound
	}
	kv := getResp.Kvs[0]
	if kv == nil || len(kv.Value) == 0 {
		return nil, worker.ErrNotFound
	}
	w, err := worker.NewFromBytes(kv.Value)
	if err != nil {
		return nil, err
	}
	w.SetType(typ)
	return w, nil
}

func (p *provider) checkWorkerIsReady(ctx context.Context, w worker.Worker) error {
	if err := p.checkWorkerFields(w); err != nil {
		return err
	}
	alive, reason, _ := p.probeOneWorkerLiveness(ctx, w)
	if !alive {
		return fmt.Errorf("%s", reason)
	}
	return nil
}

func (p *provider) checkWorkerFields(w worker.Worker) error {
	if w == nil {
		return worker.ErrNilWorker
	}
	if len(w.GetID().String()) == 0 {
		return worker.ErrNoWorkerID
	}
	if len(w.GetType().String()) == 0 {
		return worker.ErrNoWorkerType
	}
	if !w.GetType().Valid() {
		return worker.ErrInvalidWorkerType
	}
	if w.GetCreatedAt().IsZero() {
		return worker.ErrNoCreatedTime
	}
	// try to marshal
	if _, err := w.MarshalJSON(); err != nil {
		return errors.Wrap(err, worker.ErrMarshal.Error())
	}
	return nil
}

func (p *provider) listWorkers(ctx context.Context, workerTypes ...worker.Type) ([]worker.Worker, error) {
	var workers []worker.Worker
	if len(workerTypes) == 0 {
		workerTypes = worker.AllTypes
	}
	for _, typ := range workerTypes {
		typeWorkers, err := p.listWorkersByType(ctx, typ)
		if err != nil {
			return nil, err
		}
		workers = append(workers, typeWorkers...)
	}
	// remove invalid workers
	var validWorkers []worker.Worker
	for _, w := range workers {
		checkErr := p.checkWorkerIsReady(ctx, w)
		if checkErr == nil {
			validWorkers = append(validWorkers, w)
			continue
		}
		go func(w worker.Worker, reason string) {
			if deleteErr := p.deleteWorker(ctx, w); deleteErr != nil {
				p.Log.Errorf("failed to delete invalid worker while list workers, workerID: %s, typ: %s, deleteReason: %s, err: %v", w.GetID(), w.GetType(), reason, deleteErr)
			} else {
				p.Log.Infof("delete invalid worker while list workers success, workerID: %s, typ: %s, deleteReason: %s", w.GetID(), w.GetType(), reason)
			}
		}(w, checkErr.Error())
	}
	p.lock.Lock()
	p.forLeaderUse.allWorkers = make(map[worker.ID]worker.Worker, len(validWorkers))
	for _, w := range validWorkers {
		p.forLeaderUse.allWorkers[w.GetID()] = w
	}
	p.lock.Unlock()
	return validWorkers, nil
}

func (p *provider) workerListenIncomingLogicTask(ctx context.Context, w worker.Worker) {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(w.GetID())
	p.Log.Infof("worker begin listen incoming logic task, workerID: %s", w.GetID())
	defer p.Log.Infof("worker stop listen incoming logic task, workerID: %s", w.GetID())

	p.ListenPrefix(ctx, prefix, func(ctx context.Context, event *clientv3.Event) {
		go func() {
			// key added, do logic
			key := string(event.Kv.Key)
			taskLogicID := p.getWorkerTaskLogicIDFromIncomingKey(w.GetID(), key)
			taskData := event.Kv.Value
			p.Log.Infof("logic task received and begin handle it, workerID: %s, logicTaskID: %s", w.GetID(), taskLogicID)
			taskDoneCh := make(chan struct{})
			go func() {
				w.Handle(ctx, worker.NewLogicTask(taskLogicID, taskData))
				taskDoneCh <- struct{}{}
			}()
			select {
			case <-ctx.Done():
				p.Log.Warnf("task canceled, workerID: %s, logicTaskID: %s", w.GetID(), taskLogicID)
			case <-taskDoneCh:
				p.Log.Infof("task done, workerID: %s, logicTaskID: %s", w.GetID(), taskLogicID)
				// delete task key means task done
				for {
					_, err := p.EtcdClient.Delete(context.Background(), key)
					if err == nil {
						break
					}
					p.Log.Warnf("failed to delete incoming logic task key after done(auto retry), key: %s, logicTaskID: %s, err: %v", key, taskLogicID, err)
					time.Sleep(p.Cfg.Worker.Task.RetryDeleteTaskInterval)
				}
			}
		}()
	},
		nil,
	)
}

func (p *provider) workerLivenessProber(ctx context.Context) {
	p.Log.Infof("start worker liveness prober")
	defer p.Log.Infof("end worker liveness prober")

	ticker := time.NewTicker(p.Cfg.Worker.LivenessProbeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			workers, err := p.listWorkers(ctx)
			if err != nil {
				p.Log.Errorf("failed to do worker liveness probe(need retry), err: %v", err)
				continue
			}
			for _, w := range workers {
				go func(w worker.Worker) {
					alive, deadReason, err := p.probeOneWorkerLiveness(ctx, w)
					if err != nil {
						p.Log.Errorf("failed to probe worker liveness(auto retry), workerID: %s, err: %v", w.GetID(), err)
						return
					}
					if alive {
						return
					}
					// dead, delete key
					_, err = p.EtcdClient.Txn(ctx).Then(
						clientv3.OpDelete(p.makeEtcdWorkerKey(w.GetID(), worker.Candidate)),
						clientv3.OpDelete(p.makeEtcdWorkerKey(w.GetID(), worker.Official)),
					).Commit()
					if err != nil {
						p.Log.Errorf("failed to delete dead worker related keys(auto retry), workerID: %s, dead reason: %s, err: %v", w.GetID(), deadReason, err)
						return
					}
					p.Log.Infof("dead worker deleted, workerID: %s, dead reason: %s", w.GetID(), deadReason)
				}(w)
			}
		}
	}

}

// if error != nil => need retry
// false will delete key
// string is dead reason
func (p *provider) probeOneWorkerLiveness(ctx context.Context, w worker.Worker) (alive bool, reason string, needRetryErr error) {
	lastHeartbeatUnixTime, err, needRetry := p.getWorkerLastHeartbeatUnixTime(ctx, w.GetID())
	if needRetry {
		return false, "", err
	}
	if err != nil {
		return false, err.Error(), nil
	}
	// check by probe interval
	hbCfg := p.Cfg.Worker.Heartbeat
	earliestInterval := hbCfg.ReportInterval * time.Duration(hbCfg.AllowedMaxContinueLostContactTimes)
	if lastHeartbeatUnixTime > 0 {
		earliestLastProbeAt := time.Now().Add(-earliestInterval)
		actualLastProbeAt := time.Unix(lastHeartbeatUnixTime, 0)
		if actualLastProbeAt.Before(earliestLastProbeAt) {
			return false, fmt.Sprintf("liveness heartbeat failed, valid-earliest: %s, actual: %s", earliestLastProbeAt, actualLastProbeAt), nil
		}
		return true, "", nil
	}
	return true, "", nil
}

func (p *provider) workerContinueReportHeartbeat(ctx context.Context, w worker.Worker) {
	ticker := time.NewTicker(p.Cfg.Worker.Heartbeat.ReportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := p.workerOnceReportHeartbeat(ctx, w)
			if err != nil {
				p.Log.Errorf("failed to once report heartbeat(auto retry), err: %v", err)
				continue
			}
		}
	}
}

func (p *provider) workerOnceReportHeartbeat(ctx context.Context, w worker.Worker) error {
	hctx, hcancel := context.WithTimeout(ctx, p.Cfg.Worker.Heartbeat.ReportInterval)
	defer hcancel()
	alive := w.DetectHeartbeat(hctx)
	if !alive {
		return fmt.Errorf("failed to detect worker heartbeat, workerID: %s", w.GetID())
	}
	// update lastProbeAt
	nowSec := time.Now().Round(0).Unix()
	if _, err := p.EtcdClient.Put(ctx, p.makeEtcdWorkerHeartbeatKey(w.GetID()), strutil.String(nowSec)); err != nil {
		return fmt.Errorf("failed to update last heartbeat time into etcd, workerID: %s, err: %v", w.GetID(), err)
	}
	p.Log.Debugf("worker heartbeat reported, workerID: %s", w.GetID())
	return nil
}

func (p *provider) promoteCandidateWorker(ctx context.Context, w worker.Worker) {
	ticker := time.NewTicker(p.Cfg.Worker.Candidate.ThresholdToBecomeOfficial)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.SetType(worker.Official)
			if err := p.registerWorker(ctx, w, w.GetType()); err != nil {
				p.Log.Errorf("failed to promote worker to official(auto retry), workerID: %s, err: %v", w.GetID(), err)
				continue
			}
			p.Log.Infof("promote worker to official, workerID: %s", w.GetID())
			return
		}
	}
}

func (p *provider) workerHandleDelete(ctx context.Context, w worker.Worker) {
	key := p.makeEtcdWorkerKey(w.GetID(), worker.Official)
	p.ListenPrefix(context.Background(), key, nil, func(ctx context.Context, event *clientv3.Event) {
		if string(event.Kv.Key) != key {
			return
		}
		ww, ok := p.forWorkerUse.myWorkers[w.GetID()]
		if ok && ww.CancelFunc != nil {
			ww.CancelFunc()
		}
		var wg sync.WaitGroup
		for _, h := range p.forWorkerUse.workerHandlersOnWorkerDelete {
			h := h
			wg.Add(1)
			go func() {
				defer wg.Done()
				h(ctx, Event{Type: event.Type, WorkerID: w.GetID()})
			}()
		}
		wg.Wait()
	})
}

func (p *provider) workerIntervalCleanupOnDelete(ctx context.Context, ev Event) {
	// delete heartbeat key
	go func() {
		for {
			_, err := p.EtcdClient.Delete(ctx, p.makeEtcdWorkerHeartbeatKey(ev.WorkerID))
			if err == nil {
				return
			}
			p.Log.Errorf("failed to do worker interval cleanup on delete(auto retry), step: delete heartbeat key, workerID: %s, err: %v", ev.WorkerID, err)
			time.Sleep(p.Cfg.Worker.RetryInterval)
		}
	}()
	// delete dispatch key
	go func() {
		for {
			_, err := p.EtcdClient.Delete(ctx, p.makeEtcdWorkerLogicTaskListenPrefix(ev.WorkerID), clientv3.WithPrefix())
			if err == nil {
				return
			}
			p.Log.Errorf("failed to do worker interval cleanup on delete(auto retry), step: delete dispatch key, workerID: %s, err: %v", ev.WorkerID, err)
			time.Sleep(p.Cfg.Worker.RetryInterval)
		}
	}()
}

func (p *provider) listWorkerTasks(ctx context.Context, workerID worker.ID) ([]worker.LogicTask, error) {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	resp, err := p.EtcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var tasks []worker.LogicTask
	for _, kv := range resp.Kvs {
		logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, string(kv.Key))
		logicTaskData := kv.Value
		task := worker.NewLogicTask(logicTaskID, logicTaskData)
		tasks = append(tasks, task)
	}
	return tasks, nil
}
