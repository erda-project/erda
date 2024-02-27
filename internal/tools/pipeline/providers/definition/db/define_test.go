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

package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xormplus/core"

	"github.com/erda-project/erda-infra/providers/mysqlxorm/sqlite3"
	definitionpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	sourcedb "github.com/erda-project/erda/internal/tools/pipeline/providers/source/db"
)

const (
	dbSourceName = "test.db"
	mode         = "rwc"
)

func TestListPipelineDefinition(t *testing.T) {

	dbname := filepath.Join(os.TempDir(), dbSourceName)
	defer func() {
		os.Remove(dbname)
	}()
	sqlite3Db, err := sqlite3.NewSqlite3(dbname + "?mode=" + mode)
	sqlite3Db.DB().SetMapper(core.GonicMapper{})

	// migrator db
	err = sqlite3Db.DB().Sync2(&PipelineDefinition{})
	if err != nil {
		panic(err)
	}

	err = sqlite3Db.DB().Sync2(&sourcedb.PipelineSource{})
	if err != nil {
		panic(err)
	}

	client := &Client{
		Interface: sqlite3Db,
	}

	sourceClient := &sourcedb.Client{
		Interface: sqlite3Db,
	}

	// insert definition
	definitions := []*PipelineDefinition{
		{ID: "1", Name: "1", PipelineSourceId: "1", Location: "1"},
		{ID: "2", Name: "2", PipelineSourceId: "2", Location: "1"},
		{ID: "3", Name: "3", PipelineSourceId: "3", Location: "2"},
		{ID: "4", Name: "4", PipelineSourceId: "1", SoftDeletedAt: uint64(time.Now().UnixNano()), Location: "1"},
		{ID: "5", Name: "5", PipelineSourceId: "1", Location: "1"},
	}

	sources := []*sourcedb.PipelineSource{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	for _, d := range definitions {
		err = client.CreatePipelineDefinition(d)
		if err != nil {
			panic(err)
		}
	}

	for _, s := range sources {
		err = sourceClient.CreatePipelineSource(s)
		if err != nil {
			panic(err)
		}
	}

	// list definition
	ds, count, err := client.ListPipelineDefinition(&definitionpb.PipelineDefinitionListRequest{
		PageSize: int64(2),
		PageNo:   int64(1),
		Location: "1",
	})
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int64(3), count)
	assert.Equal(t, 2, len(ds))
	assert.Equal(t, "1", ds[0].ID)
	assert.Equal(t, "2", ds[1].ID)
	t.Logf("%+v", ds)

}
