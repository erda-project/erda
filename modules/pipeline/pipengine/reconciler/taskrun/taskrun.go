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

package taskrun

import (
	"context"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/throttler"
	"github.com/erda-project/erda/modules/pipeline/providers/aop/plugins_manage"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore"
)

// TaskRun represents task runtime.
type TaskRun struct {
	Task *spec.PipelineTask

	Ctx                   context.Context
	Executor              types.ActionExecutor
	Throttler             throttler.Throttler
	P                     *spec.Pipeline
	QueriedPipelineStatus apistructs.PipelineStatus

	Bdl      *bundle.Bundle
	DBClient *dbclient.Client
	Js       jsonstore.JsonStore

	QuitQueueTimeout bool
	QuitWaitTimeout  bool

	StopQueueLoop bool
	StopWaitLoop  bool

	PExitCh       <-chan struct{}
	PExitChCancel context.CancelFunc
	PExit         bool

	// 轮训状态间隔期间可能任务已经是终态，FakeTimeout = true
	FakeTimeout bool

	// svc
	ActionAgentSvc *actionagentsvc.ActionAgentSvc
	ExtMarketSvc   *extmarketsvc.ExtMarketSvc

	// aop
	PluginsManage *plugins_manage.PluginsManage
}

// New returns a TaskRun.
func New(ctx context.Context, task *spec.PipelineTask,
	pExitCh <-chan struct{}, pExitChCancel context.CancelFunc,
	throttler throttler.Throttler,
	executor types.ActionExecutor, p *spec.Pipeline, bdl *bundle.Bundle, dbClient *dbclient.Client, js jsonstore.JsonStore,
	actionAgentSvc *actionagentsvc.ActionAgentSvc,
	extMarketSvc *extmarketsvc.ExtMarketSvc,
	PluginsManage *plugins_manage.PluginsManage,
) *TaskRun {
	return &TaskRun{
		Ctx:       ctx,
		Task:      task,
		Executor:  executor,
		Throttler: throttler,
		P:         p,

		Bdl:      bdl,
		DBClient: dbClient,
		Js:       js,

		QuitQueueTimeout: false,
		QuitWaitTimeout:  false,

		StopQueueLoop: false,
		StopWaitLoop:  false,

		PExitCh:       pExitCh,
		PExitChCancel: pExitChCancel,

		ActionAgentSvc: actionAgentSvc,
		ExtMarketSvc:   extMarketSvc,

		PluginsManage: PluginsManage,
	}
}

// Op represents task operation.
type Op string

const (
	Prepare Op = "prepare"
	Create  Op = "create"
	Start   Op = "start"
	Queue   Op = "queue"
	Wait    Op = "wait"
)

type TaskOp interface {
	Op() Op

	TaskRun() *TaskRun

	// Processing represents what the `op` really do, you shouldn't update task inside `processing`.
	// you should put update logic in `WhenXXX`.
	Processing() (interface{}, error)

	// WhenDone will be invoked if task op is done.
	WhenDone(data interface{}) error

	// WhenLogicError will be invoked if task occurred an error when do logic.
	WhenLogicError(err error) error

	// WhenTimeout will be invoked if task is timeout.
	WhenTimeout() error

	TimeoutConfig() (<-chan struct{}, context.CancelFunc, time.Duration)

	// TuneTriggers return corresponding triggers at concrete tune point.
	TuneTriggers() TaskOpTuneTriggers
}

type Elem struct {
	TimeoutCh <-chan struct{}
	Cancel    context.CancelFunc
	Timeout   time.Duration

	ErrCh  chan error
	DoneCh chan interface{}

	ExitCh chan struct{}
}

type TaskOpTuneTriggers struct {
	BeforeProcessing aoptypes.TuneTrigger
	AfterProcessing  aoptypes.TuneTrigger
}
