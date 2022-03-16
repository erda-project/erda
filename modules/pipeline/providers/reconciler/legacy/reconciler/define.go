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
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pkg/action_info"
	"github.com/erda-project/erda/modules/pipeline/providers/dbgc"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
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

	// processingTasks store task id which is in processing
	processingTasks sync.Map
	// teardownPipelines store pipeline id which is in the process of tear down
	teardownPipelines sync.Map
	// processPipelines store reconciler pipeline id
	processingPipelines sync.Map

	// svc
	actionAgentSvc  *actionagentsvc.ActionAgentSvc
	extMarketSvc    *extmarketsvc.ExtMarketSvc
	pipelineSvcFunc *PipelineSvcFunc
	DBgc            dbgc.Interface
}

// In order to solve the problem of circular dependency if Reconciler introduces pipelinesvc, the svc method is mounted in this structure.
// todo resolve cycle import here through better module architecture
type PipelineSvcFunc struct {
	CronNotExecuteCompensate                func(id uint64) error
	MergePipelineYmlTasks                   func(pipelineYml *pipelineyml.PipelineYml, dbTasks []spec.PipelineTask, p *spec.Pipeline, dbStages []spec.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) (mergeTasks []spec.PipelineTask, err error)
	HandleQueryPipelineYamlBySnippetConfigs func(sourceSnippetConfigs []apistructs.SnippetConfig) (map[string]string, error)
	MakeSnippetPipeline4Create              func(p *spec.Pipeline, snippetTask *spec.PipelineTask, yamlContent string) (*spec.Pipeline, error)
	CreatePipelineGraph                     func(p *spec.Pipeline) (err error)
}

// New generate a new reconciler.
func New(js jsonstore.JsonStore, etcd *etcd.Store, bdl *bundle.Bundle, dbClient *dbclient.Client,
	actionAgentSvc *actionagentsvc.ActionAgentSvc,
	extMarketSvc *extmarketsvc.ExtMarketSvc,
	pipelineSvcFunc *PipelineSvcFunc,
	DBgc dbgc.Interface,
) (*Reconciler, error) {
	r := Reconciler{
		js:       js,
		etcd:     etcd,
		bdl:      bdl,
		dbClient: dbClient,

		processingTasks:     sync.Map{},
		teardownPipelines:   sync.Map{},
		processingPipelines: sync.Map{},

		actionAgentSvc:  actionAgentSvc,
		extMarketSvc:    extMarketSvc,
		pipelineSvcFunc: pipelineSvcFunc,
		DBgc:            DBgc,
	}
	return &r, nil
}
