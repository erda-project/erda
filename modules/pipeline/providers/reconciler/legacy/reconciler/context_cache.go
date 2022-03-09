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

// @Title  This file is used to cache the objects that need to be reused in the reconciler
// @Description  The cached data is placed in the global sync.Map, you need to call clearPipelineContextCaches to clean up the corresponding data after you use it at will
package reconciler

import (
	"fmt"
	"sync"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pkg/action_info"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

const PipelineStagesContextCachesPrefixKey = "reconciler_caches_stages"
const PipelineYmlContextCachesPrefixKey = "reconciler_caches_yml"
const PipelineRerunSuccessTaskContextCachesPrefixKey = "reconciler_caches_rerun_success_tasks"
const PipelinePassedDataWhenCreateContextCachesPrefixKey = "reconciler_caches_passed_data_when_create"

const PipelineContextCacheKey = "reconciler_caches"

var cacheMap = sync.Map{}

// clear context map rwLock by pipelineID
func clearPipelineContextCaches(pipelineID uint64) {
	cacheMap.Delete(makeMapKey(pipelineID, PipelineStagesContextCachesPrefixKey))
	cacheMap.Delete(makeMapKey(pipelineID, PipelineYmlContextCachesPrefixKey))
	cacheMap.Delete(makeMapKey(pipelineID, PipelineRerunSuccessTaskContextCachesPrefixKey))
	cacheMap.Delete(makeMapKey(pipelineID, PipelinePassedDataWhenCreateContextCachesPrefixKey))
}

// get cache map from context
func makeMapKey(pipelineID uint64, key string) string {
	return fmt.Sprintf("%v_%v_%v", pipelineID, PipelineContextCacheKey, key)
}

// obtain all types of values from the map according to the key, and then do the conversion at each call
func getInterfaceValueByKey(pipelineID uint64, key string) interface{} {
	value, ok := cacheMap.Load(makeMapKey(pipelineID, key))
	if !ok {
		return nil
	}
	return value
}

// ----------- caches pipeline stages
// get stages from context caches map by pipelineID
func getStagesCachesFromContextByPipelineID(pipelineID uint64) []spec.PipelineStage {
	value := getInterfaceValueByKey(pipelineID, PipelineStagesContextCachesPrefixKey)
	if value == nil {
		return nil
	}
	stages, ok := value.([]spec.PipelineStage)
	if !ok {
		return nil
	}

	return stages
}

func setStagesCachesToContextByPipelineID(stages []spec.PipelineStage, pipelineID uint64) {
	cacheMap.Store(makeMapKey(pipelineID, PipelineStagesContextCachesPrefixKey), stages)
}

// get value from caches map
// if not exist search value from db and save to caches map
func getOrSetStagesFromContext(dbClient *dbclient.Client, pipelineID uint64) (stages []spec.PipelineStage, err error) {
	stages = getStagesCachesFromContextByPipelineID(pipelineID)
	if stages != nil {
		return stages, nil
	}

	stages, err = dbClient.ListPipelineStageByPipelineID(pipelineID)
	if err != nil {
		return nil, err
	}

	setStagesCachesToContextByPipelineID(stages, pipelineID)
	return stages, nil
}

// -------- cache PipelineYml
func getPipelineYmlCachesFromContextByPipelineID(pipelineID uint64) *pipelineyml.PipelineYml {
	value := getInterfaceValueByKey(pipelineID, PipelineYmlContextCachesPrefixKey)
	if value == nil {
		return nil
	}
	yml, ok := value.(*pipelineyml.PipelineYml)
	if !ok {
		return nil
	}

	return yml
}

func setPipelineYmlCachesToContextByPipelineID(yml *pipelineyml.PipelineYml, pipelineID uint64) {
	cacheMap.Store(makeMapKey(pipelineID, PipelineYmlContextCachesPrefixKey), yml)
}

// get value from caches map
// if not exist search value from db and save to caches map
func getOrSetPipelineYmlFromContext(dbClient *dbclient.Client, pipelineID uint64) (yml *pipelineyml.PipelineYml, err error) {
	yml = getPipelineYmlCachesFromContextByPipelineID(pipelineID)
	if yml != nil {
		return yml, nil
	}
	pipeline, err := dbClient.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}
	pipelineYml, err := pipelineyml.New(
		[]byte(pipeline.PipelineYml),
	)
	if err != nil {
		return nil, err
	}

	setPipelineYmlCachesToContextByPipelineID(pipelineYml, pipelineID)
	return pipelineYml, nil
}

// ------- cache pre pipeline success tasks map[string]*spec.PipelineTask
func getPipelineRerunSuccessTasksFromContextByPipelineID(pipelineID uint64) map[string]*spec.PipelineTask {
	value := getInterfaceValueByKey(pipelineID, PipelineRerunSuccessTaskContextCachesPrefixKey)
	if value == nil {
		return nil
	}
	successTasks, ok := value.(map[string]*spec.PipelineTask)
	if !ok {
		return nil
	}

	return successTasks
}

func setPipelineRerunSuccessTasksToContextByPipelineID(successTasks map[string]*spec.PipelineTask, pipelineID uint64) {
	cacheMap.Store(makeMapKey(pipelineID, PipelineRerunSuccessTaskContextCachesPrefixKey), successTasks)
}

// get value from caches map
// if not exist search value from db and save to caches map
func getOrSetPipelineRerunSuccessTasksFromContext(dbClient *dbclient.Client, pipelineID uint64) (successTasks map[string]*spec.PipelineTask, err error) {
	successTasks = getPipelineRerunSuccessTasksFromContextByPipelineID(pipelineID)
	if successTasks != nil {
		return successTasks, nil
	}
	pipeline, err := dbClient.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}
	lastSuccessTaskMap, _, err := dbClient.ParseRerunFailedDetail(pipeline.Extra.RerunFailedDetail)
	if err != nil {
		return nil, err
	}
	setPipelineRerunSuccessTasksToContextByPipelineID(lastSuccessTaskMap, pipelineID)
	return lastSuccessTaskMap, nil
}

// ------- cache task_extensions.PassedDataWhenCreate
func getPassedDataWhenCreateFromContextByPipelineID(pipelineID uint64) *action_info.PassedDataWhenCreate {

	value := getInterfaceValueByKey(pipelineID, PipelinePassedDataWhenCreateContextCachesPrefixKey)
	if value == nil {
		return nil
	}
	passedData, ok := value.(*action_info.PassedDataWhenCreate)
	if !ok {
		return nil
	}

	return passedData
}

func setPassedDataWhenCreateToContextByPipelineID(passedDataWhenCreate *action_info.PassedDataWhenCreate, pipelineID uint64) {
	cacheMap.Store(makeMapKey(pipelineID, PipelinePassedDataWhenCreateContextCachesPrefixKey), passedDataWhenCreate)
}

// get value from caches map
// if not exist search value from db and save to caches map
func getOrSetPassedDataWhenCreateFromContext(bdl *bundle.Bundle, pipelineYml *pipelineyml.PipelineYml, pipelineID uint64) (passedDataWhenCreate *action_info.PassedDataWhenCreate, err error) {
	passedDataWhenCreate = getPassedDataWhenCreateFromContextByPipelineID(pipelineID)
	if passedDataWhenCreate != nil {
		return passedDataWhenCreate, nil
	}

	passedDataWhenCreate = &action_info.PassedDataWhenCreate{}
	passedDataWhenCreate.InitData(bdl)
	if err := passedDataWhenCreate.PutPassedDataByPipelineYml(pipelineYml); err != nil {
		return nil, err
	}

	setPassedDataWhenCreateToContextByPipelineID(passedDataWhenCreate, pipelineID)
	return passedDataWhenCreate, nil
}
