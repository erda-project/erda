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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/xormplus/core"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-infra/providers/mysqlxorm/sqlite3"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	crondb "github.com/erda-project/erda/internal/tools/pipeline/providers/cron/db"
	definitiondb "github.com/erda-project/erda/internal/tools/pipeline/providers/definition/db"
	sourcedb "github.com/erda-project/erda/internal/tools/pipeline/providers/source/db"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/arrays"
	mocklogger "github.com/erda-project/erda/pkg/mock"
)

// if add insert template, you should update these index
const (
	dbSourceName              = "test.db"
	mode                      = "rwc"
	sourceDeletedIndex        = 6
	sourceRepeatIndex         = 0
	sourceRepeatLength        = 3
	GroupSourceDeletedIndex   = 3
	definitionDeletedIndex    = 5
	latestExecDefinitionIndex = 4
	testTx                    = 1
	testTxBase                = 2
	testTxError               = "test tx error"
)

type MockCronService struct {
}

func (m *MockCronService) CronCreate(ctx context.Context, request *cronpb.CronCreateRequest) (*cronpb.CronCreateResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronPaging(ctx context.Context, request *cronpb.CronPagingRequest) (*cronpb.CronPagingResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronStart(ctx context.Context, request *cronpb.CronStartRequest) (*cronpb.CronStartResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronStop(ctx context.Context, request *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronDelete(ctx context.Context, request *cronpb.CronDeleteRequest) (*cronpb.CronDeleteResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronGet(ctx context.Context, request *cronpb.CronGetRequest) (*cronpb.CronGetResponse, error) {
	return nil, nil
}

func (m *MockCronService) CronUpdate(ctx context.Context, request *cronpb.CronUpdateRequest) (*cronpb.CronUpdateResponse, error) {
	return nil, nil
}

func newSqlite3DB(dbSourceName string) *sqlite3.Sqlite3 {
	sqlite3Db, err := sqlite3.NewSqlite3(dbSourceName + "?mode=" + mode)
	sqlite3Db.DB().SetMapper(core.GonicMapper{})
	if err != nil {
		panic(err)
	}

	// migrator db
	err = sqlite3Db.DB().Sync2(&definitiondb.PipelineDefinitionExtra{})
	if err != nil {
		panic(err)
	}
	err = sqlite3Db.DB().Sync2(&definitiondb.PipelineDefinition{})
	if err != nil {
		panic(err)
	}
	err = sqlite3Db.DB().Sync2(&sourcedb.PipelineSource{})
	if err != nil {
		panic(err)
	}
	err = sqlite3Db.DB().Sync2(&spec.PipelineBase{})
	if err != nil {
		panic(err)
	}
	err = sqlite3Db.DB().Sync2(&crondb.PipelineCron{})
	if err != nil {
		panic(err)
	}

	return sqlite3Db
}

func newProvider(t *testing.T, dbSourceName string, ctrl *gomock.Controller) *provider {
	sqlite3Db := newSqlite3DB(dbSourceName)

	logger := mocklogger.NewMockLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Return().AnyTimes()
	p := &provider{
		MySQL: sqlite3Db,
		Cfg:   &config{DryRun: false},
		Log:   logger,
	}

	p.Init(nil)
	p.CronService = &MockCronService{}

	return p
}

func getInsertSourceRecords() []sourcedb.PipelineSource {
	return []sourcedb.PipelineSource{
		{
			SourceType:  "erda",
			Remote:      "remote1",
			Ref:         "ref1",
			Path:        "path1",
			Name:        "name.yaml",
			PipelineYml: "",
			UpdatedAt:   time.Now().Add(1 * time.Hour),
		},
		{
			SourceType:  "erda",
			Remote:      "remote1",
			Ref:         "ref1",
			Path:        "path1",
			Name:        "name.yaml",
			PipelineYml: "",
		},
		{
			SourceType:  "erda",
			Remote:      "remote1",
			Ref:         "ref1",
			Path:        "path1",
			Name:        "name.yaml",
			PipelineYml: "",
		},
		{
			SourceType:    "erda",
			Remote:        "remote1",
			Ref:           "ref1",
			Path:          "path1",
			Name:          "name.yaml",
			PipelineYml:   "",
			SoftDeletedAt: uint64(time.Now().UnixNano() / 1e6),
		},
		{
			SourceType:  "erda",
			Remote:      "remote1",
			Ref:         "ref1",
			Path:        "path1",
			Name:        "pipeline.yaml",
			PipelineYml: "",
		},
		{
			SourceType:  "erda",
			Remote:      "remote2",
			Ref:         "ref2",
			Path:        "path2",
			Name:        "pipeline.yaml",
			PipelineYml: "",
		},
		{
			SourceType:    "erda",
			Remote:        "remote2",
			Ref:           "ref2",
			Path:          "path2",
			Name:          "pipeline.yaml",
			PipelineYml:   "",
			SoftDeletedAt: uint64(time.Now().UnixNano() / 1e6),
		},
	}
}

func getDefaultDefinitionTemplate() []definitiondb.PipelineDefinition {
	return []definitiondb.PipelineDefinition{
		{
			ID:               "1",
			Location:         "location1",
			Name:             "name1",
			Status:           "",
			Ref:              "ref1",
			PipelineSourceId: "",
		},
		{
			ID:               "2",
			Location:         "location2",
			Name:             "name2",
			Status:           "",
			Ref:              "ref1",
			PipelineSourceId: "",
		},
		{
			ID:               "3",
			Location:         "location3",
			Name:             "name3",
			Status:           "Success",
			Ref:              "ref2",
			PipelineSourceId: "",
		},
		{
			ID:               "4",
			Location:         "location4",
			Name:             "name3",
			Status:           "Success",
			Ref:              "ref2",
			PipelineSourceId: "",
			TimeCreated:      time.Now().Add(1 * time.Hour),
			TimeUpdated:      time.Now().Add(1 * time.Hour),
		},
		{
			ID:               "5",
			Location:         "location5",
			Name:             "name5",
			Status:           "Failed",
			Ref:              "ref5",
			PipelineSourceId: "",
			TimeCreated:      time.Now().Add(2 * time.Hour),
			TimeUpdated:      time.Now().Add(2 * time.Hour),
			StartedAt:        time.Now().Add(1 * time.Hour),
		},
		{
			ID:               "6",
			Location:         "location5",
			Name:             "name5",
			Status:           "Failed",
			Ref:              "ref5",
			PipelineSourceId: "",
			TimeCreated:      time.Now().Add(2 * time.Hour),
			TimeUpdated:      time.Now().Add(2 * time.Hour),
			SoftDeletedAt:    uint64(time.Now().UnixNano() / 1e6),
		},
	}
}

func getDefaultDefinitionExtraTemplate() []definitiondb.PipelineDefinitionExtra {
	return []definitiondb.PipelineDefinitionExtra{
		{ID: "1", PipelineDefinitionID: "1"},
		{ID: "2", PipelineDefinitionID: "2"},
		{ID: "3", PipelineDefinitionID: "3"},
		{ID: "4", PipelineDefinitionID: "4"},
		{ID: "5", PipelineDefinitionID: "5"},
	}
}

func getDefaultCronTemplate() []crondb.PipelineCron {
	trueFlag := true
	falseFlag := false
	return []crondb.PipelineCron{
		{
			ID:                   1,
			Enable:               &trueFlag,
			PipelineDefinitionID: "1",
		},
		{
			ID:                   2,
			Enable:               &trueFlag,
			PipelineDefinitionID: "2",
		},
		{
			ID:                   3,
			Enable:               &trueFlag,
			TimeUpdated:          time.Now().Add(1 * time.Hour),
			PipelineDefinitionID: "3",
		},
		{
			ID:                   4,
			Enable:               &falseFlag,
			PipelineDefinitionID: "4",
		},
		{
			ID:                   5,
			Enable:               &falseFlag,
			TimeUpdated:          time.Now().Add(1 * time.Hour),
			PipelineDefinitionID: "5",
		},
	}
}

func getDefaultBaseTemplate() []spec.PipelineBase {
	var cronOne uint64 = 1
	var cronTwo uint64 = 2
	var cronThree uint64 = 3
	return []spec.PipelineBase{
		{ID: 1, CronID: &cronOne, PipelineDefinitionID: "1"},
		{ID: 2, CronID: &cronOne, PipelineDefinitionID: "1"},
		{ID: 3, CronID: &cronTwo, PipelineDefinitionID: "2"},
		{ID: 4, CronID: &cronThree, PipelineDefinitionID: "3"},
		{ID: 5, CronID: nil, PipelineDefinitionID: "4"},
		{ID: 6, CronID: nil, PipelineDefinitionID: "5"},
		{ID: 7, CronID: nil, PipelineDefinitionID: "5"},
		{ID: 8, CronID: nil, PipelineDefinitionID: "5"},
	}
}

func insertSourceRecords(p *provider, insertSource []sourcedb.PipelineSource) error {
	for index, item := range insertSource {
		session := p.MySQL.NewSession()
		// no auto update time
		session.NoAutoTime()
		ops := mysqlxorm.WithSession(session)
		err := p.sourceDbClient.CreatePipelineSource(&item, ops)
		copyItem := item
		insertSource[index] = copyItem
		if err != nil {
			return err
		}
	}
	return nil
}

func insertDefinitionRecords(p *provider, insertDefinition []definitiondb.PipelineDefinition) error {
	for index, definition := range insertDefinition {
		err := p.definitionDbClient.CreatePipelineDefinition(&definition)
		copyItem := definition
		insertDefinition[index] = copyItem
		if err != nil {
			return err
		}
	}
	return nil
}

func insertCronRecords(p *provider, insertCron []crondb.PipelineCron) error {
	for index, cron := range insertCron {
		session := p.MySQL.NewSession()
		// no auto update time
		session.NoAutoTime()
		ops := mysqlxorm.WithSession(session)
		err := p.cronDbClient.CreatePipelineCron(&cron, ops)
		copyCron := cron
		insertCron[index] = copyCron
		if err != nil {
			return err
		}
	}
	return nil
}

func insertDefinitionExtraRecords(p *provider, insertDefinitionExtra []definitiondb.PipelineDefinitionExtra) error {
	for index, extra := range insertDefinitionExtra {
		err := p.definitionDbClient.CreatePipelineDefinitionExtra(&extra)
		copyExtra := extra
		insertDefinitionExtra[index] = copyExtra
		if err != nil {
			return err
		}
	}
	return nil
}

func insertBaseRecords(p *provider, insertBase []spec.PipelineBase) error {
	for index, base := range insertBase {
		err := p.dbClient.CreatePipelineBase(&base)
		copyBase := base
		insertBase[index] = copyBase
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteAllTable(p *provider) error {
	_, err := p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", sourcedb.PipelineSource{}.TableName()))
	if err != nil {
		return err
	}

	_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", definitiondb.PipelineDefinition{}.TableName()))
	if err != nil {
		return err
	}

	_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", definitiondb.PipelineDefinitionExtra{}.TableName()))
	if err != nil {
		return err
	}

	_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", crondb.PipelineCron{}.TableName()))
	if err != nil {
		return err
	}

	_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", (&spec.PipelineBase{}).TableName()))
	if err != nil {
		return err
	}

	return nil
}

func TestNeedCleanup(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := newProvider(t, dbname, ctrl)

	insertSource := getInsertSourceRecords()

	err := insertSourceRecords(p, insertSource)
	if err != nil {
		t.Fatalf("insert source err : %s", err)
	}

	want := struct {
		SourceGroupList []sourcedb.PipelineSourceUniqueGroupWithCount
		Flag            bool
	}{
		SourceGroupList: []sourcedb.PipelineSourceUniqueGroupWithCount{
			{
				SourceType: "erda",
				Remote:     "remote1",
				Ref:        "ref1",
				Path:       "path1",
				Name:       "name.yaml",
				Count:      3,
			},
			{
				SourceType: "erda",
				Remote:     "remote1",
				Ref:        "ref1",
				Path:       "path1",
				Name:       "pipeline.yaml",
				Count:      1,
			},
			{
				SourceType: "erda",
				Remote:     "remote2",
				Ref:        "ref2",
				Path:       "path2",
				Name:       "pipeline.yaml",
				Count:      1,
			},
		},
		Flag: true,
	}

	sourceList, needCleanup, err := p.needCleanup()
	if err != nil {
		t.Errorf("check need cleanup err : %s", err)
		return
	}

	assert.Equal(t, want.Flag, needCleanup)
	assert.Equal(t, want.SourceGroupList, sourceList)

	// delete repeat.
	deleteSourceList, err := p.sourceDbClient.GetPipelineSourceByUnique(&sourcedb.PipelineSourceUnique{
		SourceType: insertSource[0].SourceType,
		Remote:     insertSource[0].Remote,
		Ref:        insertSource[0].Ref,
		Path:       insertSource[0].Path,
		Name:       insertSource[0].Name,
	})

	if err != nil {
		t.Errorf("delete pipeline list err : %s", err)
		return
	}

	for i := 0; i < 2; i++ {
		deleteSource := deleteSourceList[i]
		deleteSource.SoftDeletedAt = uint64(time.Now().UnixNano() / 1e6)
		err = p.sourceDbClient.DeletePipelineSource(deleteSource.ID, &deleteSource)
		if err != nil {
			t.Errorf("delete pipeline source err : %s", err)
			return
		}
	}

	want.Flag = false
	want.SourceGroupList = nil

	sourceList, needCleanup, err = p.needCleanup()
	if err != nil {
		t.Errorf("check need cleanup err : %s", err)
		return
	}

	assert.Equal(t, want.Flag, needCleanup)
	assert.Equal(t, want.SourceGroupList, sourceList)
}

func TestGetSourceListByGroup(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	insertSource := getInsertSourceRecords()

	err := insertSourceRecords(p, insertSource)
	if err != nil {
		t.Fatalf("insert source err : %s", err)
	}

	group := sourcedb.PipelineSourceUniqueGroupWithCount{
		SourceType: insertSource[0].SourceType,
		Remote:     insertSource[0].Remote,
		Ref:        insertSource[0].Ref,
		Path:       insertSource[0].Path,
		Name:       insertSource[0].Name,
	}

	want := sourcedb.PipelineSource{
		SourceType: "erda",
		Remote:     "remote1",
		Ref:        "ref1",
		Path:       "path1",
		Name:       "name.yaml",
	}

	sourceList, err := p.GetSourceListByGroup(group)
	if err != nil {
		t.Errorf("get source list by group err : %s", err)
		return
	}

	for i := 0; i < len(sourceList); i++ {
		unique := sourcedb.PipelineSource{
			SourceType: sourceList[i].SourceType,
			Remote:     sourceList[i].Remote,
			Ref:        sourceList[i].Ref,
			Path:       sourceList[i].Path,
			Name:       sourceList[i].Name,
		}
		assert.Equal(t, want, unique)
	}
}

func TestGetLatestExecDefinitionBySourceIds(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	insertSource := getInsertSourceRecords()

	err := insertSourceRecords(p, insertSource)
	if err != nil {
		t.Fatalf("insert source err : %s", err)
	}

	// get group sourceList
	groupSourceList, err := p.sourceDbClient.GetPipelineSourceByUnique(&sourcedb.PipelineSourceUnique{
		SourceType: insertSource[sourceRepeatIndex].SourceType,
		Remote:     insertSource[sourceRepeatIndex].Remote,
		Ref:        insertSource[sourceRepeatIndex].Ref,
		Path:       insertSource[sourceRepeatIndex].Path,
		Name:       insertSource[sourceRepeatIndex].Name,
	})
	insertDefinitionList := getDefaultDefinitionTemplate()

	// binding a source for the latestExecDefinition(index = 4)
	insertDefinitionList[0].PipelineSourceId = groupSourceList[0].ID
	insertDefinitionList[1].PipelineSourceId = groupSourceList[1].ID
	insertDefinitionList[latestExecDefinitionIndex].PipelineSourceId = groupSourceList[2].ID

	// insert into the db
	err = insertDefinitionRecords(p, insertDefinitionList)
	if err != nil {
		t.Errorf("insert definition list err : %s ", err)
		return
	}

	sourceIds := arrays.GetFieldArrFromStruct(groupSourceList, func(s sourcedb.PipelineSource) string {
		return s.ID
	})

	definitionRespList, latestExecDefinition, err := p.GetLatestExecDefinitionBySourceIds(sourceIds)
	if err != nil {
		t.Fatalf("get latest exec definition by source ids err : %s", err)
	}

	want := struct {
		definitionList       []definitiondb.PipelineDefinition
		latestExecDefinition definitiondb.PipelineDefinition
	}{}

	// sqlite's millisecond precision timestamp storage and microsecond rounding can cause minor discrepancies, preventing direct comparison
	want.latestExecDefinition = insertDefinitionList[latestExecDefinitionIndex]
	want.definitionList = append(want.definitionList, insertDefinitionList[1])
	want.definitionList = append(want.definitionList, insertDefinitionList[0])

	assert.Equal(t, want.latestExecDefinition.ID, latestExecDefinition.ID)
	assert.Equal(t, want.latestExecDefinition.PipelineSourceId, latestExecDefinition.PipelineSourceId)

	assert.Equal(t, len(want.definitionList), len(definitionRespList))
	for index, definition := range want.definitionList {
		assert.Equal(t, want.definitionList[index].ID, definition.ID)
		assert.Equal(t, want.definitionList[index].PipelineSourceId, definition.PipelineSourceId)
	}
}

func TestMergeCronByDefinitionIds(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	trueFlag := true
	falseFlag := false

	insertDefinition := getDefaultDefinitionTemplate()
	err := insertDefinitionRecords(p, insertDefinition)
	if err != nil {
		t.Fatalf("insert definition records error : %s", err)
		return
	}

	definitionIds := arrays.GetFieldArrFromStruct(insertDefinition, func(d definitiondb.PipelineDefinition) string {
		return d.ID
	})
	definitionIds = append(definitionIds[:definitionDeletedIndex], definitionIds[definitionDeletedIndex+1:]...)

	type Want struct {
		latestExecDefinitionCron *crondb.PipelineCron
		cronStopIds              []uint64
		cronDbLength             int
	}

	testcase := []struct {
		name                   string
		insertPipelineCronList []crondb.PipelineCron
		definitionIds          []string
		latestExecDefinition   definitiondb.PipelineDefinition
		want                   Want
	}{
		{
			name: "test1: the cron associated with the latest definition is already started ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &falseFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &trueFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &trueFlag},
				{ID: 5, PipelineDefinitionID: "5", Enable: &trueFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   5,
					PipelineDefinitionID: "5",
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{3, 4},
				cronDbLength: 1,
			},
		},
		{
			name: "test2: the cron associated with latest definition is not started and other is not started",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &falseFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
				{ID: 5, PipelineDefinitionID: "5", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   5,
					PipelineDefinitionID: "5",
					Enable:               &falseFlag,
				},
				cronStopIds:  []uint64{},
				cronDbLength: 1,
			},
		},
		{
			name: "test3: the cron associated with latest definition is started and other is not started ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &falseFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
				{ID: 5, PipelineDefinitionID: "5", Enable: &trueFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   5,
					PipelineDefinitionID: "5",
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{},
				cronDbLength: 1,
			},
		},
		{
			name: "test4: the cron associated with latest definition is not exist and other is not started ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &falseFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: nil,
				cronStopIds:              []uint64{},
				cronDbLength:             0,
			},
		},
		{
			name: "test5: the cron associated with latest definition is not exist and the other is started(only one started) ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &trueFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &falseFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   1,
					PipelineDefinitionID: insertDefinition[latestExecDefinitionIndex].ID,
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{},
				cronDbLength: 1,
			},
		},
		{
			name: "test6: the cron associated with latest definition is not exist and the other is started(exceed 1) ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &trueFlag, TimeUpdated: time.Now()},
				{ID: 2, PipelineDefinitionID: "2", Enable: &trueFlag, TimeUpdated: time.Now().Add(1 * time.Hour)},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   2,
					PipelineDefinitionID: insertDefinition[latestExecDefinitionIndex].ID,
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{1},
				cronDbLength: 1,
			},
		},
		{
			name: "test7: the cron associated with latest definition is exist and not started, the other is started(only 1) ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &trueFlag},
				{ID: 3, PipelineDefinitionID: "3", Enable: &falseFlag},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
				{ID: 5, PipelineDefinitionID: "5", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   2,
					PipelineDefinitionID: insertDefinition[latestExecDefinitionIndex].ID,
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{},
				cronDbLength: 1,
			},
		},
		{
			name: "test8: the cron associated with latest definition is exist and not started , the other is started(exceed 1) ",
			insertPipelineCronList: []crondb.PipelineCron{
				{ID: 1, PipelineDefinitionID: "1", Enable: &falseFlag},
				{ID: 2, PipelineDefinitionID: "2", Enable: &trueFlag, TimeUpdated: time.Now().Add(-1 * time.Hour)},
				{ID: 3, PipelineDefinitionID: "3", Enable: &trueFlag, TimeUpdated: time.Now().Add(1 * time.Hour)},
				{ID: 4, PipelineDefinitionID: "4", Enable: &falseFlag},
				{ID: 5, PipelineDefinitionID: "5", Enable: &falseFlag},
			},
			definitionIds:        definitionIds,
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			want: Want{
				latestExecDefinitionCron: &crondb.PipelineCron{
					ID:                   3,
					PipelineDefinitionID: insertDefinition[latestExecDefinitionIndex].ID,
					Enable:               &trueFlag,
				},
				cronStopIds:  []uint64{2},
				cronDbLength: 1,
			},
		},
	}

	for _, test := range testcase {
		// truncate table
		t.Log(test.name)
		_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", crondb.PipelineCron{}.TableName()))
		if err != nil {
			t.Fatalf("truncate table err : %s", err)
		}

		// insert cron
		err = insertCronRecords(p, test.insertPipelineCronList)
		if err != nil {
			t.Fatalf("insert cron records err : %s", err)
		}

		cronStopIdList := []uint64{}
		// mock cronStop
		monkey.PatchInstanceMethod(reflect.TypeOf(p.CronService), "CronStop", func(_ *MockCronService, ctx context.Context, cron *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
			// get cron
			cronInfo, _, err := p.cronDbClient.GetPipelineCron(cron.CronID)
			if err != nil {
				return nil, err
			}

			if cronInfo.Enable == &trueFlag {
				return nil, err
			}

			cronInfo.Enable = &falseFlag
			err = p.cronDbClient.UpdatePipelineCron(cronInfo.ID, &cronInfo)
			if err != nil {
				return nil, err
			}
			cronStopIdList = append(cronStopIdList, cronInfo.ID)
			return &cronpb.CronStopResponse{}, nil
		})

		// check if the cron deleted
		_, latestExecCron, err := p.MergeCronByDefinitionIds(context.Background(), test.definitionIds, test.latestExecDefinition)
		if err != nil {
			t.Fatalf("merge cron by definition error")
		}

		go func() {
			monkey.UnpatchAll()
		}()

		if test.want.latestExecDefinitionCron != nil {
			assert.Equal(t, test.want.latestExecDefinitionCron.ID, latestExecCron.ID)
			assert.Equal(t, test.want.latestExecDefinitionCron.PipelineDefinitionID, latestExecCron.PipelineDefinitionID)
			assert.Equal(t, test.want.latestExecDefinitionCron.Enable, latestExecCron.Enable)
		} else {
			assert.Equal(t, test.want.latestExecDefinitionCron, latestExecCron)
		}

		// check stop cron
		sort.Slice(cronStopIdList, func(i, j int) bool {
			return cronStopIdList[i] < cronStopIdList[j]
		})
		assert.Equal(t, test.want.cronStopIds, cronStopIdList)

		// get cron record from db
		// and it must have the latestExecCron
		cronDbList := []crondb.PipelineCron{}
		err = p.MySQL.DB().Find(&cronDbList)
		if err != nil {
			t.Fatalf("get pipeline cron err : %s", err)
		}

		assert.Equal(t, test.want.cronDbLength, len(cronDbList))
		if test.want.latestExecDefinitionCron != nil {
			assert.Equal(t, test.want.latestExecDefinitionCron.ID, cronDbList[0].ID)
			assert.Equal(t, test.want.latestExecDefinitionCron.PipelineDefinitionID, cronDbList[0].PipelineDefinitionID)
			assert.Equal(t, test.want.latestExecDefinitionCron.Enable, cronDbList[0].Enable)
		} else {
			assert.Equal(t, []crondb.PipelineCron{}, cronDbList)
		}
	}
}

func TestMergeDefinition(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	insertDefinition := getDefaultDefinitionTemplate()
	insertDefinitionExtra := getDefaultDefinitionExtraTemplate()

	err := insertDefinitionRecords(p, insertDefinition)
	if err != nil {
		t.Fatalf("insert definition records error : %s", err)
	}

	err = insertDefinitionExtraRecords(p, insertDefinitionExtra)
	if err != nil {
		t.Fatalf("insert definition extra records error : %s", err)
	}

	mergeDefinitionIds := arrays.GetFieldArrFromStruct(insertDefinition, func(d definitiondb.PipelineDefinition) string {
		return d.ID
	})

	if err = p.MergeDefinition(mergeDefinitionIds); err != nil {
		t.Fatalf("merge definition err : %s", err)
	}

	definitionCount, err := p.MySQL.DB().Where("soft_deleted_at = 0").Count(&definitiondb.PipelineDefinition{})
	if err != nil {
		t.Fatalf("select count(*) from definition err : %s", err)
	}
	definitionExtraCount, err := p.MySQL.DB().Where("soft_deleted_at = 0").Count(&definitiondb.PipelineDefinitionExtra{})
	if err != nil {
		t.Fatalf("select count(*) from definition_extra err : %s", err)
	}
	assert.Equal(t, int64(0), definitionCount)
	assert.Equal(t, int64(0), definitionExtraCount)
}

func TestMergePipelineBase(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	insertDefinition := getDefaultDefinitionTemplate()
	definitionIds := arrays.GetFieldArrFromStruct(insertDefinition, func(d definitiondb.PipelineDefinition) string {
		return d.ID
	})
	definitionIds = append(definitionIds[:definitionDeletedIndex], definitionIds[definitionDeletedIndex+1:]...)

	err := insertDefinitionRecords(p, insertDefinition)
	if err != nil {
		t.Fatalf("insert definition records error : %s", err)
	}

	var latestCronId uint64 = 1
	var otherCronId uint64 = 3

	type Want struct {
		baseList []spec.PipelineBase
	}

	testCase := []struct {
		name                     string
		latestExecDefinitionCron *crondb.PipelineCron
		latestExecDefinition     definitiondb.PipelineDefinition
		definitionIds            []string
		insertBase               []spec.PipelineBase
		want                     Want
	}{
		{
			name:                     "test1: cron is not exist",
			latestExecDefinitionCron: nil,
			latestExecDefinition:     insertDefinition[latestExecDefinitionIndex],
			definitionIds:            definitionIds,
			insertBase: []spec.PipelineBase{
				{ID: 1, PipelineDefinitionID: "1", CronID: nil},
				{ID: 2, PipelineDefinitionID: "2", CronID: nil},
				{ID: 3, PipelineDefinitionID: "5", CronID: nil},
				{ID: 4, PipelineDefinitionID: "5", CronID: nil},
			},
			want: Want{
				baseList: []spec.PipelineBase{
					{ID: 1, PipelineDefinitionID: "5", CronID: nil},
					{ID: 2, PipelineDefinitionID: "5", CronID: nil},
					{ID: 3, PipelineDefinitionID: "5", CronID: nil},
					{ID: 4, PipelineDefinitionID: "5", CronID: nil},
				},
			},
		},
		{
			name: "test2: cron is exist",
			latestExecDefinitionCron: &crondb.PipelineCron{
				ID: latestCronId,
			},
			latestExecDefinition: insertDefinition[latestExecDefinitionIndex],
			definitionIds:        definitionIds,
			insertBase: []spec.PipelineBase{
				{ID: 1, PipelineDefinitionID: "1", CronID: nil},
				{ID: 2, PipelineDefinitionID: "2", CronID: &otherCronId},
				{ID: 3, PipelineDefinitionID: "5", CronID: &otherCronId},
				{ID: 4, PipelineDefinitionID: "5", CronID: &latestCronId},
			},
			want: Want{
				baseList: []spec.PipelineBase{
					{ID: 1, PipelineDefinitionID: "5", CronID: nil},
					{ID: 2, PipelineDefinitionID: "5", CronID: &latestCronId},
					{ID: 3, PipelineDefinitionID: "5", CronID: &latestCronId},
					{ID: 4, PipelineDefinitionID: "5", CronID: &latestCronId},
				},
			},
		},
	}

	for _, test := range testCase {
		t.Log(test.name)

		_, err = p.MySQL.DB().Exec(fmt.Sprintf("DELETE FROM %s", (&spec.PipelineBase{}).TableName()))
		if err != nil {
			t.Fatalf("truncate table err : %s", err)
		}

		// insert base
		err = insertBaseRecords(p, test.insertBase)
		if err != nil {
			t.Fatalf("insert base records err : %s", err)
		}

		err = p.MergePipelineBase(test.definitionIds, test.latestExecDefinition, test.latestExecDefinitionCron)
		if err != nil {
			t.Fatalf("merge pipeline base err : %s", err)
		}

		// select * from pipeline_bases
		baseInDbList := []spec.PipelineBase{}
		err = p.MySQL.DB().Find(&baseInDbList)

		// sort by id
		sort.Slice(baseInDbList, func(i, j int) bool {
			return baseInDbList[i].ID < baseInDbList[j].ID
		})

		assert.Equal(t, len(test.want.baseList), len(baseInDbList))
		for index, base := range baseInDbList {
			assert.Equal(t, test.want.baseList[index].ID, base.ID)
			assert.Equal(t, test.want.baseList[index].CronID, base.CronID)
			assert.Equal(t, test.want.baseList[index].PipelineDefinitionID, base.PipelineDefinitionID)
		}
	}

}

func TestMergeSource(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	testCase := []struct {
		name                 string
		latestExecDefinition definitiondb.PipelineDefinition
		saveSourceIndex      int
	}{
		{name: "latestExecDefinition is not null", latestExecDefinition: definitiondb.PipelineDefinition{ID: "1"}, saveSourceIndex: 1},
		{name: "latestExecDefinition is null", latestExecDefinition: definitiondb.PipelineDefinition{ID: ""}, saveSourceIndex: 0},
	}

	for _, test := range testCase {
		err := deleteAllTable(p)
		if err != nil {
			t.Fatalf("delete all table err: %s", err)
		}

		insertSourceList := getInsertSourceRecords()
		err = insertSourceRecords(p, insertSourceList)
		if err != nil {
			t.Fatalf("insert source records err: %s", err)
		}
		insertSourceList = insertSourceList[:sourceDeletedIndex]

		test.latestExecDefinition.PipelineSourceId = insertSourceList[test.saveSourceIndex].ID

		err = p.MergeSource(insertSourceList, test.latestExecDefinition)
		if err != nil {
			t.Fatalf("merge source err : %s", err)
		}

		savePipelineSourec := insertSourceList[test.saveSourceIndex]
		pipelineSourceList := []sourcedb.PipelineSource{}
		p.MySQL.DB().Where("soft_deleted_at = 0").Find(&pipelineSourceList)

		assert.Equal(t, 1, len(pipelineSourceList))
		assert.Equal(t, savePipelineSourec.ID, pipelineSourceList[0].ID)

	}

	//insertSourceList := getInsertSourceRecords()
	//err := insertSourceRecords(p, insertSourceList)
	//if err != nil {
	//	t.Fatalf("insert source records err : %s", err)
	//}
	//insertSourceList = insertSourceList[:sourceDeletedIndex]
	//
	//// latestExecDefinition is not null
	//latestExecDefinition := definitiondb.PipelineDefinition{
	//	ID:               "1",
	//	PipelineSourceId: insertSourceList[0].ID,
	//}
	//
	//err = p.MergeSource(insertSourceList, latestExecDefinition)
	//if err != nil {
	//	t.Fatalf("merge source err : %s", err)
	//}
	//
	//savedPipelineSource := insertSourceList[0]
	//pipelineSourceList := []sourcedb.PipelineSource{}
	//p.MySQL.DB().Where("soft_deleted_at = 0").Find(&pipelineSourceList)
	//
	//assert.Equal(t, 1, len(pipelineSourceList))
	//assert.Equal(t, savedPipelineSource.ID, pipelineSourceList[0].ID)

}

func TestMergePipeline(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := newProvider(t, dbname, ctrl)

	trueFlag := true
	//falseFlag := false

	var cronOne uint64 = 1
	//var cronTwo uint64 = 2
	var cronThree uint64 = 3

	type Want struct {
		wantSource          []sourcedb.PipelineSource
		wantDefinition      []definitiondb.PipelineDefinition
		wantDefinitionExtra []definitiondb.PipelineDefinitionExtra
		wantCron            []crondb.PipelineCron
		wantBase            []spec.PipelineBase
	}

	testCase := []struct {
		name                  string
		testType              int
		insertSource          []sourcedb.PipelineSource
		insertDefinition      []definitiondb.PipelineDefinition
		insertCron            []crondb.PipelineCron
		insertDefinitionExtra []definitiondb.PipelineDefinitionExtra
		insertBase            []spec.PipelineBase
		uniqueGroup           sourcedb.PipelineSourceUniqueGroupWithCount
		want                  Want
	}{
		{
			name:                  "test1: success",
			insertSource:          getInsertSourceRecords(),
			insertDefinition:      getDefaultDefinitionTemplate(),
			insertDefinitionExtra: getDefaultDefinitionExtraTemplate(),
			insertCron:            getDefaultCronTemplate(),
			insertBase:            getDefaultBaseTemplate(),
			uniqueGroup: sourcedb.PipelineSourceUniqueGroupWithCount{
				SourceType: getInsertSourceRecords()[0].SourceType,
				Remote:     getInsertSourceRecords()[0].Remote,
				Ref:        getInsertSourceRecords()[0].Ref,
				Path:       getInsertSourceRecords()[0].Path,
				Name:       getInsertSourceRecords()[0].Name,
			},
			want: Want{
				wantSource: []sourcedb.PipelineSource{
					{SourceType: getInsertSourceRecords()[0].SourceType, Remote: getInsertSourceRecords()[0].Remote, Ref: getInsertSourceRecords()[0].Ref, Path: getInsertSourceRecords()[0].Path, Name: getInsertSourceRecords()[0].Name},
					{SourceType: "erda", Remote: "remote1", Ref: "ref1", Path: "path1", Name: "pipeline.yaml", PipelineYml: ""},
					{SourceType: "erda", Remote: "remote2", Ref: "ref2", Path: "path2", Name: "pipeline.yaml", PipelineYml: ""},
				},
				wantDefinition: []definitiondb.PipelineDefinition{
					{
						ID:               "1",
						Location:         "location1",
						Name:             "name1",
						Status:           "",
						Ref:              "ref1",
						PipelineSourceId: "",
					},
					{
						ID:               "5",
						Location:         "location5",
						Name:             "name5",
						Status:           "Failed",
						Ref:              "ref5",
						PipelineSourceId: "",
						TimeCreated:      time.Now().Add(2 * time.Hour),
						TimeUpdated:      time.Now().Add(2 * time.Hour),
					},
				},
				wantDefinitionExtra: []definitiondb.PipelineDefinitionExtra{
					{ID: "1", PipelineDefinitionID: "1"},
					{ID: "5", PipelineDefinitionID: "5"},
				},
				wantCron: []crondb.PipelineCron{
					{
						ID:                   1,
						Enable:               &trueFlag,
						PipelineDefinitionID: "1",
					},
					{
						ID:                   3,
						Enable:               &trueFlag,
						TimeUpdated:          time.Now().Add(1 * time.Hour),
						PipelineDefinitionID: "5",
					},
				},
				wantBase: []spec.PipelineBase{
					{ID: 1, CronID: &cronOne, PipelineDefinitionID: "1"},
					{ID: 2, CronID: &cronOne, PipelineDefinitionID: "1"},
					{ID: 3, CronID: &cronThree, PipelineDefinitionID: "5"},
					{ID: 4, CronID: &cronThree, PipelineDefinitionID: "5"},
					{ID: 5, CronID: nil, PipelineDefinitionID: "5"},
					{ID: 6, CronID: nil, PipelineDefinitionID: "5"},
					{ID: 7, CronID: nil, PipelineDefinitionID: "5"},
					{ID: 8, CronID: nil, PipelineDefinitionID: "5"},
				},
			},
		},
		{
			name:                  "test2: test tx error, mock MergeSource",
			testType:              testTx,
			insertSource:          getInsertSourceRecords(),
			insertDefinition:      getDefaultDefinitionTemplate(),
			insertDefinitionExtra: getDefaultDefinitionExtraTemplate(),
			insertCron:            getDefaultCronTemplate(),
			insertBase:            getDefaultBaseTemplate(),
			uniqueGroup: sourcedb.PipelineSourceUniqueGroupWithCount{
				SourceType: getInsertSourceRecords()[0].SourceType,
				Remote:     getInsertSourceRecords()[0].Remote,
				Ref:        getInsertSourceRecords()[0].Ref,
				Path:       getInsertSourceRecords()[0].Path,
				Name:       getInsertSourceRecords()[0].Name,
			},
		},
		{
			name:                  "test3: test tx error, mock MergeBase",
			testType:              testTxBase,
			insertSource:          getInsertSourceRecords(),
			insertDefinition:      getDefaultDefinitionTemplate(),
			insertDefinitionExtra: getDefaultDefinitionExtraTemplate(),
			insertCron:            getDefaultCronTemplate(),
			insertBase:            getDefaultBaseTemplate(),
			uniqueGroup: sourcedb.PipelineSourceUniqueGroupWithCount{
				SourceType: getInsertSourceRecords()[0].SourceType,
				Remote:     getInsertSourceRecords()[0].Remote,
				Ref:        getInsertSourceRecords()[0].Ref,
				Path:       getInsertSourceRecords()[0].Path,
				Name:       getInsertSourceRecords()[0].Name,
			},
		},
	}

	for _, test := range testCase {
		t.Log(test.name)

		err := deleteAllTable(p)
		if err != nil {
			t.Fatalf("delete all table err : %s", err)
		}

		// insert records
		err = insertSourceRecords(p, test.insertSource)
		if err != nil {
			t.Fatalf("insert source records err : %s", err)
		}

		test.insertDefinition[0].PipelineSourceId = test.insertSource[3].ID
		test.insertDefinition[1].PipelineSourceId = test.insertSource[0].ID
		test.insertDefinition[2].PipelineSourceId = test.insertSource[2].ID
		test.insertDefinition[3].PipelineSourceId = test.insertSource[1].ID
		test.insertDefinition[latestExecDefinitionIndex].PipelineSourceId = test.insertSource[0].ID

		for i := 0; i < len(test.want.wantDefinition); i++ {
			for _, definition := range test.insertDefinition {
				if definition.ID == test.want.wantDefinition[i].ID {
					test.want.wantDefinition[i].PipelineSourceId = definition.PipelineSourceId
					break
				}
			}
		}

		err = insertDefinitionRecords(p, test.insertDefinition)
		if err != nil {
			t.Fatalf("insert definition records err : %s", err)
		}

		err = insertDefinitionExtraRecords(p, test.insertDefinitionExtra)
		if err != nil {
			t.Fatalf("insert definition extra err : %s", err)
		}

		err = insertCronRecords(p, test.insertCron)
		if err != nil {
			t.Fatalf("insert cron records err : %s", err)
		}

		err = insertBaseRecords(p, test.insertBase)
		if err != nil {
			t.Fatalf("insert base records err : %s", err)
		}

		if test.testType == testTx {
			monkey.PatchInstanceMethod(reflect.TypeOf(p), "MergeSource", func(_ *provider, sourceList []sourcedb.PipelineSource, latestExecDefinition definitiondb.PipelineDefinition, ops ...mysqlxorm.SessionOption) error {
				return errors.New(testTxError)
			})
		}

		if test.testType == testTxBase {
			monkey.PatchInstanceMethod(reflect.TypeOf(p), "MergePipelineBase", func(_ *provider, definitionIds []string, latestExecDefinition definitiondb.PipelineDefinition, latestExecDefinitionCron *crondb.PipelineCron, ops ...dbclient.SessionOption) (err error) {
				return errors.New(testTxError)
			})
		}

		if test.testType != 0 {
			// set want as insert
			for _, source := range test.insertSource {
				if source.SoftDeletedAt == 0 {
					test.want.wantSource = append(test.want.wantSource, source)
				}
			}

			for _, definition := range test.insertDefinition {
				if definition.SoftDeletedAt == 0 {
					test.want.wantDefinition = append(test.want.wantDefinition, definition)
				}
			}

			for _, extra := range test.insertDefinitionExtra {
				if extra.SoftDeletedAt == 0 {
					test.want.wantDefinitionExtra = append(test.want.wantDefinitionExtra, extra)
				}
			}

			test.want.wantCron = test.insertCron
			test.want.wantBase = test.insertBase
			sort.Slice(test.want.wantCron, func(i, j int) bool {
				return test.want.wantCron[i].ID < test.want.wantCron[j].ID
			})
		}

		err = p.MergePipeline(context.Background(), test.uniqueGroup)
		if err != nil && err.Error() != testTxError {
			t.Fatalf("merge pipeline err : %s", err)
		}

		sourceList := []sourcedb.PipelineSource{}
		definitionList := []definitiondb.PipelineDefinition{}
		definitionExtraList := []definitiondb.PipelineDefinitionExtra{}
		cronList := []crondb.PipelineCron{}
		baseList := []spec.PipelineBase{}

		// get records in db
		err = p.MySQL.DB().Where("soft_deleted_at = 0").Find(&sourceList)
		if err != nil {
			t.Fatalf("get source list err : %s", err)
		}

		err = p.MySQL.DB().Where("soft_deleted_at = 0").Find(&definitionList)
		if err != nil {
			t.Fatalf("get definition list err : %s", err)
		}

		err = p.MySQL.DB().Where("soft_deleted_at = 0").Find(&definitionExtraList)
		if err != nil {
			t.Fatalf("get definition extra list err : %s", err)
		}

		err = p.MySQL.DB().Find(&cronList)
		if err != nil {
			t.Fatalf("get cron list err : %s", err)
		}

		err = p.MySQL.DB().Find(&baseList)
		if err != nil {
			t.Fatalf("get base list err : %s", err)
		}

		// sort
		sort.Slice(sourceList, func(i, j int) bool {
			return sourceList[i].CreatedAt.Before(sourceList[j].CreatedAt)
		})
		sort.Slice(definitionList, func(i, j int) bool {
			return definitionList[i].ID < definitionList[j].ID
		})
		sort.Slice(definitionExtraList, func(i, j int) bool {
			return definitionExtraList[i].ID < definitionExtraList[j].ID
		})
		sort.Slice(cronList, func(i, j int) bool {
			return cronList[i].ID < cronList[j].ID
		})
		sort.Slice(baseList, func(i, j int) bool {
			return baseList[i].ID < baseList[j].ID
		})

		assert.Equal(t, len(test.want.wantSource), len(sourceList))
		assert.Equal(t, len(test.want.wantDefinition), len(definitionList))
		assert.Equal(t, len(test.want.wantDefinitionExtra), len(definitionExtraList))
		assert.Equal(t, len(test.want.wantCron), len(cronList))
		assert.Equal(t, len(test.want.wantBase), len(baseList))

		for index, source := range sourceList {
			assert.Equal(t, test.want.wantSource[index].Ref, source.Ref)
			assert.Equal(t, test.want.wantSource[index].Remote, source.Remote)
			assert.Equal(t, test.want.wantSource[index].SourceType, source.SourceType)
			assert.Equal(t, test.want.wantSource[index].Path, source.Path)
			assert.Equal(t, test.want.wantSource[index].Name, source.Name)
		}

		for index, definition := range definitionList {
			assert.Equal(t, test.want.wantDefinition[index].ID, definition.ID)
			assert.Equal(t, test.want.wantDefinition[index].Location, definition.Location)
			assert.Equal(t, test.want.wantDefinition[index].Name, definition.Name)
			assert.Equal(t, test.want.wantDefinition[index].Status, definition.Status)
			assert.Equal(t, test.want.wantDefinition[index].Ref, definition.Ref)
			assert.Equal(t, test.want.wantDefinition[index].PipelineSourceId, definition.PipelineSourceId)
		}

		for index, extra := range definitionExtraList {
			assert.Equal(t, test.want.wantDefinitionExtra[index].ID, extra.ID)
			assert.Equal(t, test.want.wantDefinitionExtra[index].PipelineDefinitionID, extra.PipelineDefinitionID)
		}

		for index, cron := range cronList {
			assert.Equal(t, test.want.wantCron[index].ID, cron.ID)
			assert.Equal(t, test.want.wantCron[index].PipelineDefinitionID, cron.PipelineDefinitionID)
			assert.Equal(t, test.want.wantCron[index].Enable, cron.Enable)
		}

		for index, base := range baseList {
			assert.Equal(t, test.want.wantBase[index].ID, base.ID)
			assert.Equal(t, test.want.wantBase[index].CronID, base.CronID)
			assert.Equal(t, test.want.wantBase[index].PipelineDefinitionID, base.PipelineDefinitionID)
		}
	}
}
