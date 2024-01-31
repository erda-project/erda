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

package definition_cleanup

import (
	"context"
	"errors"
	"time"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	crondb "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/db"
	definitiondb "github.com/erda-project/erda/internal/tools/pipeline/providers/definition/db"
	sourcedb "github.com/erda-project/erda/internal/tools/pipeline/providers/source/db"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/arrays"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/retry"
)

/*
	This method is used to clean up and merge duplicate pipeline source/definition/definition_extra/cron/base records

	Design Reason：
		Due to previous issue (https://github.com/erda-project/erda/pull/6170) with the pipeline creation logic,
		there are a large number of duplicate records in the db that stores pipeline records.
        After fixing the logic issues, it is necessary to clean up and merge these duplicate records.
	Merge Principle：
		The latest exec definition(the one with the latest `StartAt` and non-deleted) and its corresponding source records will be the merge baseline.
        All other related records will be merged into this records baseline (definition, source, definition_extra, cron will be deleted, and bases will be associated with the retained definition record)
    Execution Time：
		Since it is a long transaction, we want to minimize the impact on user operations. Therefore, it is scheduled to run at midnight once a day.
    More Information：
		Normally, most of the records will be cleaned up in the first execution. There may be some leftovers that will be cleaned up in the next cleanup step.
        After all, there should be no more duplicate records, therefore, in subsequent runs, this job mainly serves as a check function.

	Attention:
		This method may still cause a few db record relationship issues in extremes cases
		(
			eg.
				A new base record linked to post-merge deleted data in a single tx results in data loss after the tx committing,
				making the base record could not get from the definition.
		)
		however, since all operations in this step are soft-deletes（`excluding pipeline_cron`）, the records can retrieved again, if in error cases, you should recovery it in manual

	For more detail information, please refer to the code and comments.
*/

const dryrunPrifix = "pipeline-definition-cleanup"

// RepeatPipelineRecordCleanup
func (p *provider) RepeatPipelineRecordCleanup(ctx context.Context) {
	cronProcess := cron.New()
	err := cronProcess.AddFunc(p.Cfg.CronExpr, func() {
		p.cronCleanup(ctx)
	})
	if err != nil {
		panic(err)
	}
	cronProcess.Start()
}

func (p *provider) cronCleanup(ctx context.Context) {
	uniqueSourceList, needCleanup, err := p.needCleanup()
	if err != nil {
		p.Log.Errorf("check need cleanup error: %s", err)
		return
	}

	if p.Cfg.DryRun {
		p.Log.Infof("\nget SourceList length is: %d, and needCleanup is %t", len(uniqueSourceList), needCleanup)
	}

	if !needCleanup {
		return
	}
	p.doCleanupRepeatRecords(ctx, uniqueSourceList)

	return
}

func (p *provider) doCleanupRepeatRecords(ctx context.Context, uniqueSourceGroupList []sourcedb.PipelineSourceUniqueGroupWithCount) {
	p.Log.Info("[Start Cleanup]")
	for index, group := range uniqueSourceGroupList {
		// if the uniqueSourceGroup.count is not equal 1, it means it has the repeat records, need merge and cleanup
		// otherwise, skip the clean logic
		if group.Count != 1 {
			var err error
			p.Log.Infof("Start this group cleanup index: %d", index+1)
			err = retry.DoWithInterval(func() error {
				err = p.MergePipeline(ctx, group)
				return err
			}, p.Cfg.RetryTimes, p.Cfg.RetryInterval)

			if err != nil {
				p.Log.Errorf("merge pipeline by source definition error: %s", err)
				return
			}

			// Add some sleep time, yield the CPU
			time.Sleep(2 * time.Second)
		}
	}
}

// needCleanup check if the db should be cleanup
// if need cleanup, it returns the `uniqueSourceList`, true and err
// otherwise, it returns nil, false and err
// This method has several scenarios：
//  1. between the time interval of querying `GetUniqueSourceGroup` and `CountPipelineSource`, if some source records is deleted
//     it may result in `count` < `len(uniqueSourceList)`. in this case, it is impossible to determine whether there are duplicate records.
//     however, since the possibility of duplicate records is low (it they occur, it indicates a problem pipeline creation logic).
//     therefore, we can wait for the next cleaning operation
//  2. if `count` > `len(uniqueSourceList)`, it is impossible to determine whether there are duplicate records or if some source records was added during the two query process
//     in this case, we directly iterate and perform a cleanup
//  3. if they are equal, it is considered that there are no duplicate records
//     so we skip the cleaning process this time
func (p *provider) needCleanup() ([]sourcedb.PipelineSourceUniqueGroupWithCount, bool, error) {
	uniqueSourceList, err := p.sourceDbClient.GetUniqueSourceGroup()
	if err != nil {
		return nil, false, err
	}
	count, err := p.sourceDbClient.CountPipelineSource()

	if int(count) > len(uniqueSourceList) {
		return uniqueSourceList, true, nil
	}

	return nil, false, nil
}

// MergePipeline
func (p *provider) MergePipeline(ctx context.Context, uniqueGroup sourcedb.PipelineSourceUniqueGroupWithCount) (err error) {
	if p.Cfg.DryRun {
		p.Log.Infof("[Current Cleanup Group]: \n %+v", uniqueGroup)
	}

	// start a transaction to do this task
	tx := p.MySQL.NewSession()
	defer tx.Close()

	if err = tx.Begin(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rbErr := tx.Rollback()
			if p.Cfg.DryRun {
				p.Log.Infof("[Result]: current cleanup group:\n %+v rollback", uniqueGroup)
			}
			if rbErr != nil {
				p.Log.Errorf("failed to rollback mergePipeline failed,uniqueSourceGroup: %v , rollbackErr: %v", uniqueGroup, rbErr)
			}
			return
		}
		cmErr := tx.Commit()
		if p.Cfg.DryRun {
			p.Log.Infof("[Result]: current cleanup group:\n %+v commit", uniqueGroup)
		}
		if cmErr != nil {
			p.Log.Errorf("failed to commit when mergePipeline success, uniqueSourceGroup: %v, commitErr: %v", uniqueGroup, cmErr)
		}
	}()

	ops := mysqlxorm.WithSession(tx)

	// 1. get repeated source list by unique
	sourceList, err := p.GetSourceListByGroup(uniqueGroup, ops)
	if err != nil {
		p.Log.Errorf("get pipeline source err: %s", err)
		return err
	}

	// construct the sourceIds from sourceList
	sourceIds := arrays.GetFieldArrFromStruct(sourceList, func(s sourcedb.PipelineSource) string {
		return s.ID
	})

	if p.Cfg.DryRun {
		p.Log.Infof("[Get Source List By Group] (length : %d): \n sourceIds: %v", len(sourceIds), sourceIds)
	}

	// 2. get definition list by sourceIds order by created_at
	definitionList, latestExecDefinition, err := p.GetLatestExecDefinitionBySourceIds(sourceIds, ops)
	if err != nil {
		p.Log.Errorf("get pipeline definition err: %s", err)
		return err
	}

	if len(definitionList) <= 0 {
		p.Log.Errorf("definition is less 1")
		// merge source directly
		err = p.MergeSource(sourceList, latestExecDefinition, ops)
		if err != nil {
			p.Log.Errorf("merge source err: %s", err)
		}
		return err
	}

	// construct definition ids from definitionList
	definitionIds := arrays.GetFieldArrFromStruct(definitionList, func(d definitiondb.PipelineDefinition) string {
		return d.ID
	})

	// 3. get cron by definition id which enabled is false and delete it
	_, latestExecDefinitionCron, err := p.MergeCronByDefinitionIds(ctx, definitionIds, latestExecDefinition, ops)
	if err != nil {
		p.Log.Errorf("merge pipeline cron err: %s", err)
		return err
	}

	// 4. delete definition and definition_extra
	err = p.MergeDefinition(definitionIds, ops)
	if err != nil {
		p.Log.Errorf("batch delete pipeline definition and extra err: %s", err)
		return err
	}

	// 5. merge pipeline base
	err = p.MergePipelineBase(definitionIds, latestExecDefinition, latestExecDefinitionCron, dbclient.WithTxSession(tx.Session))
	if err != nil {
		p.Log.Errorf("merge pipeline base err: %s", err)
		return err
	}

	// 6. merge source in the latest step
	err = p.MergeSource(sourceList, latestExecDefinition, ops)
	if err != nil {
		p.Log.Errorf("merge source err: %s", err)
		return err
	}

	return nil
}

func (p *provider) GetSourceListByGroup(uniqueGroup sourcedb.PipelineSourceUniqueGroupWithCount, ops ...mysqlxorm.SessionOption) (sourceList []sourcedb.PipelineSource, err error) {
	sourceList, err = p.sourceDbClient.GetPipelineSourceByUnique(&sourcedb.PipelineSourceUnique{
		SourceType: uniqueGroup.SourceType,
		Remote:     uniqueGroup.Remote,
		Ref:        uniqueGroup.Ref,
		Path:       uniqueGroup.Path,
		Name:       uniqueGroup.Name,
	}, ops...)

	if err != nil {
		return nil, err
	}

	// if source length is less 1, returns
	if len(sourceList) <= 1 {
		return nil, errors.New("the source length is less 1")
	}

	return
}

// GetLatestExecDefinitionBySourceIds returns definitionList by sourceIds which filter the latestExecDefinition, latestExecDefinition, error
// get leastExecDefinition by `StartedAt`
func (p *provider) GetLatestExecDefinitionBySourceIds(sourceIds []string, ops ...mysqlxorm.SessionOption) (definitionList []definitiondb.PipelineDefinition, latestExecDefinition definitiondb.PipelineDefinition, err error) {
	definitionList, err = p.definitionDbClient.GetPipelineDefinitionListInSourceIDs(sourceIds, ops...)
	if err != nil {
		return nil, definitiondb.PipelineDefinition{}, err
	}

	if len(definitionList) <= 0 {
		return nil, definitiondb.PipelineDefinition{}, nil
	}

	latestExecDefinition = definitionList[0]
	latestStart := definitionList[0].StartedAt
	latestIndex := 0

	// get the latest exec definition based on the start time as the saved records
	// the latest execution is determined by the definition which has the latest start time
	for index, definition := range definitionList {
		if definition.StartedAt.After(latestStart) {
			latestStart = definition.StartedAt
			latestExecDefinition = definition
			latestIndex = index
		}
	}

	definitionList = append(definitionList[:latestIndex], definitionList[latestIndex+1:]...)

	if p.Cfg.DryRun {
		p.Log.Infof("[Get Definition List By Source Besides LatestExecDefinition] definitionList (length : %d): \n", len(definitionList))
		for _, d := range definitionList {
			p.Log.Infof("definition: { id : %s, sourceId : %s, startAt: %v, timeCreate: %v }", d.ID, d.PipelineSourceId, d.StartedAt, d.TimeCreated)
		}
		p.Log.Infof("[LatestExecDefinition]\n { id: %s, sourceId: %s, startAt: %v, timeCreate: %v}", latestExecDefinition.ID, latestExecDefinition.PipelineSourceId, latestExecDefinition.StartedAt, latestExecDefinition.TimeCreated)
	}

	return
}

// MergeCronByDefinitionIds
// The principle of merging here is to avoid losing all the started timed pipelines
// There are several situations:
// 1. if the cron corresponding to the saved definition is already started, other cron pipelines should be paused and deleted
// 2. if all the other cron records(besides the cron binding the saved definition) are not started, delete them directly
// 3. if the cron corresponding to the saved definition is not started, find the latest updated cron which has started as the saved cron record (modify its binding definition ID), if not have the started cron, do not change
// 4. if the saved definition does not have a cron record, select the latest updated cron which has started as the saved cron record, if not have the started cron, d not change
func (p *provider) MergeCronByDefinitionIds(ctx context.Context, definitionIds []string, latestExecDefinition definitiondb.PipelineDefinition, ops ...mysqlxorm.SessionOption) (cronList []crondb.PipelineCron, latestExecDefinitionCron *crondb.PipelineCron, err error) {
	definitionIds = append(definitionIds, latestExecDefinition.ID)
	cronList, err = p.cronDbClient.BatchGetPipelineCronByDefinitionID(definitionIds, ops...)
	if err != nil {
		return nil, nil, err
	}

	// due to the cron is directly deleted, this prints out all the fields of all the cron found in the search
	if len(cronList) > 0 {
		p.Log.Infof("[Batch Get Cron By Definition ID %d] (length : %d): \n", latestExecDefinition.ID, len(cronList))
		for _, c := range cronList {
			p.Log.Infof("cron is: %+v，enable: %t", c, *c.Enable)
		}
	}

	if p.Cfg.DryRun {
		if len(cronList) > 0 {
			p.Log.Infof("[Get Cron List By DefinitionIds] (length : %d) \n", len(cronList))
		}
		for _, c := range cronList {
			p.Log.Infof("cron: { id: %d, timeUpdated: %v, DefinitionID: %s, ymlName: %s, timeCreated: %v, enable: %t }", c.ID, c.TimeUpdated, c.PipelineDefinitionID, c.PipelineYmlName, c.TimeCreated, *c.Enable)
		}
	}

	if len(cronList) <= 0 {
		return
	}

	// flag to indicate whether there is a started cron in cronList which to be deleted
	deleteCronIds := make([]uint64, 0)
	cronStartList := make([]crondb.PipelineCron, 0)

	for _, cron := range cronList {
		if cron.PipelineDefinitionID == latestExecDefinition.ID {
			copyCron := cron
			latestExecDefinitionCron = &copyCron
		} else if cron.Enable != nil && *cron.Enable {
			cronStartList = append(cronStartList, cron)
		}
		deleteCronIds = append(deleteCronIds, cron.ID)
	}

	if len(cronStartList) > 0 {
		// if there are other opened and preserved definitions where the corresponding cron is also enabled, directly stop them
		if !(latestExecDefinitionCron != nil && latestExecDefinitionCron.Enable != nil && *latestExecDefinitionCron.Enable) {
			// if it has the other cron has started, select the newestCron(sort by time_updated) as the latestExecDefinitionCron
			bindingCron := cronStartList[0]
			// update its binding definitionID
			if p.Cfg.DryRun {
				p.Log.Infof("[Update Binding Cron]\n cron: { id : %d, timeCreated : %v, timeUpdated : %v, enable: %v, defID : %s, expr : %s }", bindingCron.ID, bindingCron.TimeUpdated, bindingCron.TimeUpdated, bindingCron.Enable, bindingCron.PipelineDefinitionID, bindingCron.CronExpr)
			} else {
				err = p.cronDbClient.UpdatePipelineCron(bindingCron.ID, &crondb.PipelineCron{
					PipelineDefinitionID: latestExecDefinition.ID,
				}, ops...)

				if err != nil {
					return nil, nil, err
				}
			}

			bindingCron.PipelineDefinitionID = latestExecDefinition.ID
			latestExecDefinitionCron = &bindingCron
		}

		for _, cron := range cronStartList {
			if cron.ID == latestExecDefinitionCron.ID {
				continue
			}
			if p.Cfg.DryRun {
				p.Log.Infof("[Cron Stop] cronId : %d", cron.ID)
				continue
			}
			_, err = p.CronService.CronStop(ctx, &cronpb.CronStopRequest{
				CronID: cron.ID,
			})
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if latestExecDefinitionCron != nil {
		for index, cronId := range deleteCronIds {
			if cronId == latestExecDefinitionCron.ID {
				deleteCronIds = append(deleteCronIds[:index], deleteCronIds[index+1:]...)
				break
			}
		}
	}

	if p.Cfg.DryRun {
		// dry run for delete cron
		p.Log.Infof("[Batch Delete Pipeline Cron By ids] (length : %d) \n deleteCronIds: %+v", len(deleteCronIds), deleteCronIds)

		if latestExecDefinitionCron != nil {
			p.Log.Infof("[LatestExecDefinitionCron]: \n { id : %d, timeUpdate : %v, defID : %s, enable: %t }", latestExecDefinitionCron.ID, latestExecDefinitionCron.TimeUpdated, latestExecDefinitionCron.PipelineDefinitionID, *latestExecDefinitionCron.Enable)
		}

		return cronList, latestExecDefinitionCron, nil
	}
	err = p.cronDbClient.BatchDeletePipelineCron(deleteCronIds, ops...)
	if err != nil {
		return nil, nil, err
	}

	return
}

// MergeDefinition delete definition and extra
func (p *provider) MergeDefinition(definitionIds []string, ops ...mysqlxorm.SessionOption) (err error) {
	// batch delete definition
	if p.Cfg.DryRun {
		p.Log.Infof("[Batch Delete Pipeline Definition and extra](length : %d):\n definitionids: %+v", len(definitionIds), definitionIds)
	} else {
		err = p.definitionDbClient.BatchDeletePipelineDefinition(definitionIds, ops...)
		if err != nil {
			return err
		}
	}

	if p.Cfg.DryRun {
		return nil
	}

	// batch delete definition_extra
	err = p.definitionDbClient.BatchDeletePipelineDefinitionExtra(definitionIds, ops...)
	if err != nil {
		return err
	}

	return nil
}

// MergePipelineBase update pipeline_base's definitionId as the latestExecDefinition.ID which associate with the deleted pipeline_definition
// if cronList.length > 0, we should update the cronId if cronID exists
// otherwise, update the definitionID
func (p *provider) MergePipelineBase(definitionIds []string, latestExecDefinition definitiondb.PipelineDefinition, latestExecDefinitionCron *crondb.PipelineCron, ops ...dbclient.SessionOption) (err error) {
	// merge pipeline base has 2 steps
	// the one is update the pipeline_base.cronID
	// the other is update the pipeline_base.definitionID
	if latestExecDefinitionCron != nil {
		baseList, err := p.dbClient.GetPipelineBaseByFilter(&dbclient.PipelineBaseFilter{PipelineDefinitionID: definitionIds}, ops...)
		if err != nil {
			return err
		}

		for _, base := range baseList {
			base.PipelineDefinitionID = latestExecDefinition.ID

			if base.CronID != nil {
				base.CronID = &latestExecDefinitionCron.ID
			}

			if p.Cfg.DryRun {
				if base.CronID == nil {
					p.Log.Infof("[Update Base List When Cron exist]: base:\n { id: %d, cronID: nil, defID: %s }", base.ID, base.PipelineDefinitionID)
				} else {
					p.Log.Infof("[Update Base List When Cron exist]: base:\n { id %d, cronID %d, defID: %s }", base.ID, *base.CronID, base.PipelineDefinitionID)
				}
				continue
			}

			err = p.dbClient.UpdatePipelineBase(base.ID, &base, ops...)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if p.Cfg.DryRun {
		p.Log.Infof("[Update Base List By DefinitionIds When Cron not exist] (length : %d)\n definitionids : %+v", len(definitionIds), definitionIds)
		return nil
	}

	err = p.dbClient.BatchUpdatePipelineBaseByDefinitionIDs(definitionIds, map[spec.Field]interface{}{
		spec.FieldPipelineDefinitionID: latestExecDefinition.ID,
	}, ops...)

	if err != nil {
		return err
	}

	return nil
}

// MergeSource
// save the source records which is associated latestExecDefinition record and soft delete the others
func (p *provider) MergeSource(sourceList []sourcedb.PipelineSource, latestExecDefinition definitiondb.PipelineDefinition, ops ...mysqlxorm.SessionOption) error {
	var savedSource *sourcedb.PipelineSource
	if latestExecDefinition.ID == "" {
		// if it has no definition
		// sort by update_time
		for _, source := range sourceList {
			copySource := source
			if savedSource == nil {
				savedSource = &copySource
				continue
			}
			if source.UpdatedAt.After(savedSource.UpdatedAt) {
				savedSource = &copySource
			}
		}
	}

	for _, source := range sourceList {
		if savedSource == nil && source.ID == latestExecDefinition.PipelineSourceId {
			continue
		}
		if savedSource != nil && source.ID == savedSource.ID {
			continue
		}
		source.SoftDeletedAt = uint64(time.Now().UnixNano() / 1e6)

		if p.Cfg.DryRun {
			p.Log.Infof("[Merge Source By SourceIds]: source\n { id: %s, deleteAt: %d }", source.ID, source.SoftDeletedAt)
			continue
		}

		err := p.sourceDbClient.DeletePipelineSource(source.ID, &source, ops...)
		if err != nil {
			return err
		}
	}

	if p.Cfg.DryRun {
		p.Log.Infof("[Merge Source Length] : %d", len(sourceList))
	}

	return nil
}
