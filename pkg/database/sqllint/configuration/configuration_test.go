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

package configuration_test

import (
	"io/ioutil"
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint/configuration"
)

func TestFromData(t *testing.T) {
	var filename = "../testdata/config.yaml"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	c, err := configuration.FromData(data)
	if err != nil {
		t.Fatal(err)
	}
	assert(c, t)
}

func TestFromLocal(t *testing.T) {
	var filename = "../testdata/config.yaml"
	c, err := configuration.FromLocal(filename)
	if err != nil {
		t.Fatal(err)
	}
	assert(c, t)
}

func TestConfiguration_ToJsonIndent(t *testing.T) {
	c := new(configuration.Configuration)
	c.AllowedDDLs = new(configuration.AllowedDDLs)
	c.AllowedDMLs = new(configuration.AllowedDMLs)
	_, err := c.ToJsonIndent()
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfiguration_ToYaml(t *testing.T) {
	c := new(configuration.Configuration)
	c.AllowedDDLs = new(configuration.AllowedDDLs)
	c.AllowedDMLs = new(configuration.AllowedDMLs)
	_, err := c.ToYaml()
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfiguration_Rulers(t *testing.T) {
	c := new(configuration.Configuration)
	if _, err := c.Rulers(); err == nil {
		t.Fatal("err should not be nil")
	}
	c.AllowedDDLs = new(configuration.AllowedDDLs)
	if _, err := c.Rulers(); err == nil {
		t.Fatal("err should not be nil")
	}
	c.AllowedDMLs = new(configuration.AllowedDMLs)

	c.AllowedDDLs.CreateDatabaseStmt = true
	c.AllowedDDLs.AlterDatabaseStmt = true
	c.AllowedDDLs.DropDatabaseStmt = true
	c.AllowedDDLs.CreateTableStmt = true
	c.AllowedDDLs.DropTableStmt = true
	c.AllowedDDLs.DropSequenceStmt = true
	c.AllowedDDLs.RenameTableStmt = true
	c.AllowedDDLs.CreateViewStmt = true
	c.AllowedDDLs.CreateSequenceStmt = true
	c.AllowedDDLs.CreateIndexStmt = true
	c.AllowedDDLs.DropIndexStmt = true
	c.AllowedDDLs.LockTablesStmt = true
	c.AllowedDDLs.UnlockTablesStmt = true
	c.AllowedDDLs.CleanupTableLockStmt = true
	c.AllowedDDLs.RepairTableStmt = true
	c.AllowedDDLs.TruncateTableStmt = true
	c.AllowedDDLs.RecoverTableStmt = true
	c.AllowedDDLs.FlashBackTableStmt = true
	c.AllowedDDLs.AlterTableOption = true
	c.AllowedDDLs.AlterTableAddColumns = true
	c.AllowedDDLs.AlterTableAddConstraint = true
	c.AllowedDDLs.AlterTableDropColumn = true
	c.AllowedDDLs.AlterTableDropPrimaryKey = true
	c.AllowedDDLs.AlterTableDropIndex = true
	c.AllowedDDLs.AlterTableDropForeignKey = true
	c.AllowedDDLs.AlterTableModifyColumn = true
	c.AllowedDDLs.AlterTableChangeColumn = true
	c.AllowedDDLs.AlterTableRenameColumn = true
	c.AllowedDDLs.AlterTableRenameTable = true
	c.AllowedDDLs.AlterTableAlterColumn = true
	c.AllowedDDLs.AlterTableLock = true
	c.AllowedDDLs.AlterTableAlgorithm = true
	c.AllowedDDLs.AlterTableRenameIndex = true
	c.AllowedDDLs.AlterTableForce = true
	c.AllowedDDLs.AlterTableAddPartitions = true
	c.AllowedDDLs.AlterTableCoalescePartitions = true
	c.AllowedDDLs.AlterTableDropPartition = true
	c.AllowedDDLs.AlterTableTruncatePartition = true
	c.AllowedDDLs.AlterTablePartition = true
	c.AllowedDDLs.AlterTableEnableKeys = true
	c.AllowedDDLs.AlterTableDisableKeys = true
	c.AllowedDDLs.AlterTableRemovePartitioning = true
	c.AllowedDDLs.AlterTableWithValidation = true
	c.AllowedDDLs.AlterTableWithoutValidation = true
	c.AllowedDDLs.AlterTableSecondaryLoad = true
	c.AllowedDDLs.AlterTableSecondaryUnload = true
	c.AllowedDDLs.AlterTableRebuildPartition = true
	c.AllowedDDLs.AlterTableReorganizePartition = true
	c.AllowedDDLs.AlterTableCheckPartitions = true
	c.AllowedDDLs.AlterTableExchangePartition = true
	c.AllowedDDLs.AlterTableOptimizePartition = true
	c.AllowedDDLs.AlterTableRepairPartition = true
	c.AllowedDDLs.AlterTableImportPartitionTablespace = true
	c.AllowedDDLs.AlterTableDiscardPartitionTablespace = true
	c.AllowedDDLs.AlterTableAlterCheck = true
	c.AllowedDDLs.AlterTableDropCheck = true
	c.AllowedDDLs.AlterTableImportTablespace = true
	c.AllowedDDLs.AlterTableDiscardTablespace = true
	c.AllowedDDLs.AlterTableIndexInvisible = true
	c.AllowedDDLs.AlterTableOrderByColumns = true
	c.AllowedDDLs.AlterTableSetTiFlashReplica = true
	c.AllowedDMLs.SelectStmt = true
	c.AllowedDMLs.UnionStmt = true
	c.AllowedDMLs.LoadDataStmt = true
	c.AllowedDMLs.InsertStmt = true
	c.AllowedDMLs.DeleteStmt = true
	c.AllowedDMLs.UpdateStmt = true
	c.AllowedDMLs.ShowStmt = true
	c.AllowedDMLs.SplitRegionStmt = true
	c.BooleanFieldLinter = true
	c.CharsetLinter = true
	c.ColumnCommentLinter = true
	c.ColumnNameLinter = true
	c.CreatedAtDefaultValueLinter = true
	c.CreatedAtExistsLinter = true
	c.CreatedAtTypeLinter = true
	c.DestructLinter = true
	c.FloatDoubleLinter = true
	c.ForeignKeyLinter = true
	c.IDExistsLinter = true
	c.IDIsPrimaryLinter = true
	c.IDTypeLinter = true
	c.IndexLengthLinter = true
	c.IndexNameLinter = true
	c.KeywordsLinter = true
	c.NotNullLinter = true
	c.TableCommentLinter = true
	c.TableNameLinter = true
	c.UpdatedAtExistsLinter = true
	c.UpdatedAtTypeLinter = true
	c.UpdatedAtDefaultValueLinter = true
	c.UpdatedAtOnUpdateLinter = true
	c.VarcharLengthLinter = true
	c.CompleteInsertLinter = true
	c.ManualTimeSetterLinter = true

	if _, err := c.Rulers(); err != nil {
		t.Fatal(err)
	}
}

func assert(c *configuration.Configuration, t *testing.T) {
	if c.AllowedDDLs == nil {
		t.Fatal("failed to FromLocal, c.AllowedDDL should not be nil")
	}
	if c.AllowedDMLs == nil {
		t.Fatal("failed to FromLocal, c.AllowedDML should not be nil")
	}
	if !c.BooleanFieldLinter {
		t.Fatal("failed to FromLocal")
	}
}
