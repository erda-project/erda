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

package manager

import (
	"sync"

	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
)

// defaultManager is the default manager.
type defaultManager struct {
	queueByID         map[string]types.Queue   // key: pq id
	queueStopChanByID map[string]chan struct{} // key: pq id
	qLock             sync.RWMutex

	//pipelineCaches map[uint64]*spec.Pipeline
	//pCacheLock     sync.RWMutex

	dbClient *dbclient.Client

	// aop
	pluginsManage *plugins_manage.PluginsManage
}

// New return a new queue manager.
func New(pluginsManage *plugins_manage.PluginsManage, ops ...Option) types.QueueManager {
	var mgr defaultManager

	mgr.queueByID = make(map[string]types.Queue)
	mgr.queueStopChanByID = make(map[string]chan struct{})
	mgr.pluginsManage = pluginsManage
	//mgr.pipelineCaches = make(map[uint64]*spec.Pipeline)

	// apply options
	for _, op := range ops {
		op(&mgr)
	}

	return &mgr
}

type Option func(manager *defaultManager)

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(mgr *defaultManager) {
		mgr.dbClient = dbClient
	}
}
