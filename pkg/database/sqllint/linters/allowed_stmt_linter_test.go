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

package linters_test

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
)

const (
	allowedStmtLinter_CreateDatabaseStmt = `create database some_database;`
	// allowedStmtLinter_AlterDatabaseStmt                    = `ALTER DATABASE mydb READ ONLY = 1;`
	allowedStmtLinter_DropDatabaseStmt                     = `drop database some_database;`
	allowedStmtLinter_CrateTableStmt                       = `create table some_table (	id bigint);`
	allowedStmtLinter_DropTableStmt                        = `drop table some_table;`
	allowedStmtLinter_DropSequenceStmt                     = `DROP SEQUENCE oe.customers_seq;`
	allowedStmtLinter_RenameTableStmt                      = `rename table some_table to any_table;`
	allowedStmtLinter_CreateViewStmt                       = `CREATE VIEW test.v AS SELECT * FROM t;`
	allowedStmtLinter_CreateSequenceStmt                   = `CREATE SEQUENCE customers_seq START WITH 1000 INCREMENT BY 1 NOCACHE NOCYCLE;`
	allowedStmtLinter_CreateIndexStmt                      = `CREATE INDEX part_of_name ON customer (name(10));`
	allowedStmtLinter_DropIndexStmt                        = "DROP INDEX `PRIMARY` ON t;"
	allowedStmtLinter_LockTablesStmt                       = `LOCK TABLES t1 WRITE, t2 READ;`
	allowedStmtLinter_UnlockTablesStmt                     = `UNLOCK TABLES;`
	allowedStmtLinter_CleanupTableLockStmt                 = `ADMIN CLEANUP TABLE LOCK some_lock;`
	allowedStmtLinter_RepairTableStmt                      = `repair table t1 quick;`
	allowedStmtLinter_TruncateTableStmt                    = `truncate table some_table;`
	allowedStmtLinter_RecoverTableStmt                     = `recover table some_table;`
	allowedStmtLinter_FlashBackTableStmt                   = `flashback table some_table;`
	allowedStmtLinter_AlterTableOption                     = `alter table some_table comment 'some comment';`
	allowedStmtLinter_AlterTableAddColumns                 = `alter table some_table add column id bigint;`
	allowedStmtLinter_AlterTableAddConstraint              = `ALTER TABLE table_name ADD CONSTRAINT MyUniqueConstraint UNIQUE(column1, column2);`
	allowedStmtLinter_AlterTableDropColumn                 = `ALTER TABLE table_name DROP COLUMN column_name;`
	allowedStmtLinter_AlterTableDropPrimaryKey             = `ALTER TABLE table_name DROP PRIMARY KEY;`
	allowedStmtLinter_AlterTableDropIndex                  = `alter table some drop index my_index;`
	allowedStmtLinter_AlterTableDropForeignKey             = `alter table some drop foreign key fk_sym;`
	allowedStmtLinter_AlterTableModifyColumn               = `ALTER TABLE t1 MODIFY b INT NOT NULL;`
	allowedStmtLinter_AlterTableChangeColumn               = `ALTER TABLE t1 CHANGE a b BIGINT NOT NULL;`
	allowedStmtLinter_AlterTableRenameColumn               = `ALTER TABLE t1 RENAME COLUMN b TO a;`
	allowedStmtLinter_AlterTableRenameTable                = `alter table t1 rename to t2;`
	allowedStmtLinter_AlterTableAlterColumn                = `alter table t2 alter column c1 set default 0;`
	allowedStmtLinter_AlterTableLock                       = `alter table t1 lock none;`
	allowedStmtLinter_AlterTableAlgorithm                  = `alter table t1 algorithm default;`
	allowedStmtLinter_AlterTableRenameIndex                = `alter table t1 rename index idx_1 to idx_2;`
	allowedStmtLinter_AlterTableForce                      = `alter table t1 force;`
	allowedStmtLinter_AlterTableAddPartitions              = `alter table t1 ADD PARTITION (PARTITION P1 VALUES LESS THAN (2010));`
	allowedStmtLinter_AlterTableCoalescePartitions         = `ALTER TABLE t2 COALESCE PARTITION 2;`
	allowedStmtLinter_AlterTableDropPartition              = `ALTER TABLE table_name DROP PARTITION partition_name;`
	allowedStmtLinter_AlterTableTruncatePartition          = `ALTER TABLE table_name TRUNCATE PARTITION partition_name;`
	allowedStmtLinter_AlterTablePartition                  = `alter table t1 partition by hash(id) partitions 8;`
	allowedStmtLinter_AlterTableEnableKeys                 = `alter table t1 enable keys;`
	allowedStmtLinter_AlterTableDisableKeys                = `alter table t1 disable keys;`
	allowedStmtLinter_AlterTableRemovePartitioning         = `alter table t1 remove partitioning;`
	allowedStmtLinter_AlterTableWithValidation             = `ALTER TABLE t1 EXCHANGE PARTITION p1 WITH TABLE t2 WITH VALIDATION;`
	allowedStmtLinter_AlterTableWithoutValidation          = `ALTER TABLE target_table EXCHANGE PARTITION target_partition WITH TABLE source_table WITHOUT VALIDATION;`
	allowedStmtLinter_AlterTableSecondaryLoad              = `ALTER TABLE orders SECONDARY_ENGINE = RAPID;`
	allowedStmtLinter_AlterTableSecondaryUnload            = ``
	allowedStmtLinter_AlterTableRebuildPartition           = `ALTER TABLE t1 REBUILD PARTITION p0, p1;`
	allowedStmtLinter_AlterTableReorganizePartition        = `ALTER TABLE t1 ALGORITHM=INPLACE, REORGANIZE PARTITION;`
	allowedStmtLinter_AlterTableCheckPartitions            = `ALTER TABLE t1 CHECK PARTITION p2;`
	allowedStmtLinter_AlterTableExchangePartition          = `alter table t1 exchange partition p1 with table t2;`
	allowedStmtLinter_AlterTableOptimizePartition          = `alter table t1 optimize partition all;`
	allowedStmtLinter_AlterTableRepairPartition            = `alter table t1 repair partition all;`
	allowedStmtLinter_AlterTableImportPartitionTablespace  = `alter table t1 import partition all tablespace;`
	allowedStmtLinter_AlterTableDiscardPartitionTablespace = `alter table t1 discard partition all tablespace;`
	allowedStmtLinter_AlterTableAlterCheck                 = `alter table t1 alter check c1 enforced;`
	allowedStmtLinter_AlterTableDropCheck                  = `alter table t1 drop check c1;`
	allowedStmtLinter_AlterTableImportTablespace           = `alter table t1 import tablespace;`
	allowedStmtLinter_AlterTableDiscardTablespace          = `alter table t1 discard tablespace;`
	allowedStmtLinter_AlterTableIndexInvisible             = `ALTER TABLE t1 ALTER INDEX i_idx VISIBLE;`
	allowedStmtLinter_AlterTableOrderByColumns             = `alter table t1 order by c1, c2;`
	allowedStmtLinter_AlterTableSetTiFlashReplica          = ``
	allowedStmtLinter_SelectStmt                           = `select * from t1;`
	allowedStmtLinter_UnionStmt                            = `(select * from t1) union (select * from t2);`
	allowedStmtLinter_LoadDataStmt                         = `LOAD DATA INFILE 'data.txt' INTO TABLE db2.my_table;`
	allowedStmtLinter_InsertStmt                           = `insert into t1 values ('a', 'b');`
	allowedStmtLinter_DeleteStmt                           = `delete from t1;`
	allowedStmtLinter_UpdateStmt                           = `update t1 set c1 = c1+1;`
	allowedStmtLinter_ShowStmt                             = `show tables;`
	allowedStmtLinter_SplitRegionStmt                      = ``
)

const (
	begin  = "begin;"
	commit = "commit;"
)

func TestNewAllowedStmtLinter(t *testing.T) {
	sqlsList := []string{
		allowedStmtLinter_CreateDatabaseStmt,
		// allowedStmtLinter_AlterDatabaseStmt,
		allowedStmtLinter_DropDatabaseStmt,
		allowedStmtLinter_CrateTableStmt,
		allowedStmtLinter_DropTableStmt,
		allowedStmtLinter_DropSequenceStmt,
		allowedStmtLinter_RenameTableStmt,
		allowedStmtLinter_CreateViewStmt,
		allowedStmtLinter_CreateSequenceStmt,
		allowedStmtLinter_CreateIndexStmt,
		allowedStmtLinter_DropIndexStmt,
		allowedStmtLinter_LockTablesStmt,
		allowedStmtLinter_UnlockTablesStmt,
		allowedStmtLinter_CleanupTableLockStmt,
		// allowedStmtLinter_RepairTableStmt,
		allowedStmtLinter_TruncateTableStmt,
		allowedStmtLinter_RecoverTableStmt,
		allowedStmtLinter_FlashBackTableStmt,
		allowedStmtLinter_AlterTableOption,
		allowedStmtLinter_AlterTableAddColumns,
		allowedStmtLinter_AlterTableAddConstraint,
		allowedStmtLinter_AlterTableDropColumn,
		allowedStmtLinter_AlterTableDropPrimaryKey,
		allowedStmtLinter_AlterTableDropIndex,
		allowedStmtLinter_AlterTableDropForeignKey,
		allowedStmtLinter_AlterTableModifyColumn,
		allowedStmtLinter_AlterTableChangeColumn,
		allowedStmtLinter_AlterTableRenameColumn,
		allowedStmtLinter_AlterTableRenameTable,
		allowedStmtLinter_AlterTableAlterColumn,
		allowedStmtLinter_AlterTableLock,
		allowedStmtLinter_AlterTableAlgorithm,
		allowedStmtLinter_AlterTableRenameIndex,
		allowedStmtLinter_AlterTableForce,
		allowedStmtLinter_AlterTableAddPartitions,
		allowedStmtLinter_AlterTableCoalescePartitions,
		allowedStmtLinter_AlterTableDropPartition,
		allowedStmtLinter_AlterTableTruncatePartition,
		allowedStmtLinter_AlterTablePartition,
		allowedStmtLinter_AlterTableEnableKeys,
		allowedStmtLinter_AlterTableDisableKeys,
		allowedStmtLinter_AlterTableRemovePartitioning,
		allowedStmtLinter_AlterTableWithValidation,
		allowedStmtLinter_AlterTableWithoutValidation,
		allowedStmtLinter_AlterTableSecondaryLoad,
		allowedStmtLinter_AlterTableSecondaryUnload,
		allowedStmtLinter_AlterTableRebuildPartition,
		allowedStmtLinter_AlterTableReorganizePartition,
		allowedStmtLinter_AlterTableCheckPartitions,
		allowedStmtLinter_AlterTableExchangePartition,
		allowedStmtLinter_AlterTableOptimizePartition,
		allowedStmtLinter_AlterTableRepairPartition,
		allowedStmtLinter_AlterTableImportPartitionTablespace,
		allowedStmtLinter_AlterTableDiscardPartitionTablespace,
		allowedStmtLinter_AlterTableAlterCheck,
		allowedStmtLinter_AlterTableDropCheck,
		allowedStmtLinter_AlterTableImportTablespace,
		allowedStmtLinter_AlterTableDiscardTablespace,
		allowedStmtLinter_AlterTableIndexInvisible,
		allowedStmtLinter_AlterTableOrderByColumns,
		allowedStmtLinter_AlterTableSetTiFlashReplica,
		allowedStmtLinter_SelectStmt,
		allowedStmtLinter_UnionStmt,
		allowedStmtLinter_LoadDataStmt,
		allowedStmtLinter_InsertStmt,
		allowedStmtLinter_DeleteStmt,
		allowedStmtLinter_UpdateStmt,
		allowedStmtLinter_ShowStmt,
		allowedStmtLinter_SplitRegionStmt,
	}

	linterA := sqllint.New(linters.NewAllowedStmtLinter(linters.DDLStmt, linters.DMLStmt|linters.UpdateStmt))
	linterB := sqllint.New(linters.NewAllowedStmtLinter(0, 0))
	for _, stmt := range sqlsList {
		if err := linterA.Input([]byte(stmt), "stmt"); err != nil {
			t.Fatal(err)
		}
		if err := linterB.Input([]byte(stmt), "stmt"); err != nil {
			t.Fatal(err)
		}
	}
	if errors := linterA.Errors(); len(errors["stmt [lints]"]) > 0 {
		data, _ := json.Marshal(errors)
		t.Logf("%s", string(data))
		t.Fatal("fails")
	}
	if errors := linterB.Errors(); len(errors["stmt [lints]"]) == 0 {
		t.Fatal("fails")
	}

	if err := linterA.Input([]byte(begin), "begin"); err != nil {
		t.Fatal(err)
	}
	if errors := linterA.Errors(); len(errors["begin [lints]"]) == 0 {
		t.Fatal("fails")
	}
	data, _ := json.Marshal(linterA.Errors())
	t.Log("report:", linterA.Report(), "errors:", string(data))
}
