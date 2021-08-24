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

package pipelinesvc

import (
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// passedDataWhenCreate stores data passed recursively when create graph.
type passedDataWhenCreate struct {
	extMarketSvc     *extmarketsvc.ExtMarketSvc
	actionJobDefines *sync.Map
	actionJobSpecs   *sync.Map
}

func (that *passedDataWhenCreate) getActionJobDefine(actionTypeVersion string) *diceyml.Job {
	if that == nil {
		return nil
	}
	if that.actionJobDefines == nil {
		return nil
	}

	if value, ok := that.actionJobDefines.Load(actionTypeVersion); ok {
		if job, ok := value.(*diceyml.Job); ok {
			return job
		}
	}
	return nil
}

func (that *passedDataWhenCreate) getActionJobSpecs(actionTypeVersion string) *apistructs.ActionSpec {

	if that == nil {
		return nil
	}
	if that.actionJobDefines == nil {
		return nil
	}

	if value, ok := that.actionJobSpecs.Load(actionTypeVersion); ok {
		if spec, ok := value.(*apistructs.ActionSpec); ok {
			return spec
		}
	}
	return nil
}

func (that *passedDataWhenCreate) initData(extMarketSvc *extmarketsvc.ExtMarketSvc) {
	if that == nil {
		return
	}

	if that.actionJobDefines == nil {
		that.actionJobDefines = &sync.Map{}
	}
	if that.actionJobSpecs == nil {
		that.actionJobSpecs = &sync.Map{}
	}
	that.extMarketSvc = extMarketSvc
}

func (that *passedDataWhenCreate) putPassedDataByPipelineYml(pipelineYml *pipelineyml.PipelineYml) error {
	if that == nil {
		return nil
	}
	// batch search extensions
	var extItems []string
	for _, stage := range pipelineYml.Spec().Stages {
		for _, typedAction := range stage.Actions {
			for _, action := range typedAction {
				if action.Type.IsSnippet() {
					continue
				}
				extItem := extmarketsvc.MakeActionTypeVersion(action)
				// extension already searched, skip
				if _, ok := that.actionJobDefines.Load(extItem); ok {
					continue
				}
				extItems = append(extItems, extmarketsvc.MakeActionTypeVersion(action))
			}
		}
	}

	extItems = strutil.DedupSlice(extItems, true)
	actionJobDefines, actionJobSpecs, err := that.extMarketSvc.SearchActions(extItems)
	if err != nil {
		return apierrors.ErrCreatePipelineGraph.InternalError(err)
	}

	for extItem, actionJobDefine := range actionJobDefines {
		that.actionJobDefines.Store(extItem, actionJobDefine)
	}
	for extItem, actionJobSpec := range actionJobSpecs {
		that.actionJobSpecs.Store(extItem, actionJobSpec)
	}
	return nil
}
