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
	"fmt"
	"sync"
	"time"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/throttler"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/loop"
)

const (
	etcdReconcilerWatchPrefix = "/devops/pipeline/reconciler/"
	etcdReconcilerDLockPrefix = "/devops/pipeline/dlock/reconciler/"
	EtcdNeedCompensatePrefix  = "/devops/pipeline/compensate/"

	ctxKeyPipelineID               = "pipelineID"
	ctxKeyPipelineExitCh           = "pExitCh"
	ctxKeyPipelineExitChCancelFunc = "pExitChCancelFunc"
)

type Reconciler struct {
	js       jsonstore.JsonStore
	etcd     *etcd.Store
	bdl      *bundle.Bundle
	dbClient *dbclient.Client

	QueueManager  types.QueueManager
	TaskThrottler throttler.Throttler

	// processingTasks store task id which is in processing
	processingTasks sync.Map
	// teardownPipelines store pipeline id which is in the process of tear down
	teardownPipelines sync.Map

	// svc
	actionAgentSvc  *actionagentsvc.ActionAgentSvc
	extMarketSvc    *extmarketsvc.ExtMarketSvc
	pipelineSvcFunc *PipelineSvcFunc

	// aop
	PluginsManage *plugins_manage.PluginsManage
}

// 该结构体为了解决假如 Reconciler 引入 pipelinesvc 导致循环依赖问题，所以将 svc 方法挂载进来
type PipelineSvcFunc struct {
	CronNotExecuteCompensate func(id uint64) error
}

// New generate a new reconciler.
func New(js jsonstore.JsonStore, etcd *etcd.Store, bdl *bundle.Bundle, dbClient *dbclient.Client,
	actionAgentSvc *actionagentsvc.ActionAgentSvc,
	extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineSvcFunc *PipelineSvcFunc,
	pluginsManage *plugins_manage.PluginsManage,
) (*Reconciler, error) {
	r := Reconciler{
		js:       js,
		etcd:     etcd,
		bdl:      bdl,
		dbClient: dbClient,

		processingTasks:   sync.Map{},
		teardownPipelines: sync.Map{},

		actionAgentSvc:  actionAgentSvc,
		extMarketSvc:    extMarketSvc,
		pipelineSvcFunc: pipelineSvcFunc,
		PluginsManage: pluginsManage,
	}
	if err := r.loadThrottler(); err != nil {
		return nil, err
	}
	if err := r.loadQueueManger(); err != nil {
		return nil, err
	}
	return &r, nil
}

// Add add pipelineID to reconciler, until add success
func (r *Reconciler) Add(pipelineID uint64) {
	rlog.PInfof(pipelineID, "start add to reconciler")
	defer rlog.PInfof(pipelineID, "end add to reconciler")
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*60)).Do(func() (abort bool, err error) {
		err = r.js.Put(context.Background(), fmt.Sprintf("%s%d", etcdReconcilerWatchPrefix, pipelineID), nil)
		if err != nil {
			rlog.PErrorf(pipelineID, "add to reconciler failed, err: %v, try again later", err)
			return false, err
		}
		rlog.PInfof(pipelineID, "add to reconciler success")
		return true, nil
	})
}
