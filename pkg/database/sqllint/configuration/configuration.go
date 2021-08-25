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

package configuration

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/database/sqllint/linters"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
)

// Configuration is the structure of config yaml/json file
type Configuration struct {
	AllowedDDLs                 *AllowedDDLs `json:"allowed_ddl" yaml:"allowed_ddl"`
	AllowedDMLs                 *AllowedDMLs `json:"allowed_dml" yaml:"allowed_dml"`
	BooleanFieldLinter          bool         `json:"boolean_field_linter" yaml:"boolean_field_linter"`
	CharsetLinter               bool         `json:"charset_linter" yaml:"charset_linter"`
	ColumnCommentLinter         bool         `json:"column_comment_linter" yaml:"column_comment_linter"`
	ColumnNameLinter            bool         `json:"column_name_linter" yaml:"column_name_linter"`
	CreatedAtDefaultValueLinter bool         `json:"created_at_default_value_linter" yaml:"created_at_default_value_linter"`
	CreatedAtExistsLinter       bool         `json:"created_at_exists_linter" yaml:"created_at_exists_linter"`
	CreatedAtTypeLinter         bool         `json:"created_at_type_linter" yaml:"created_at_type_linter"`
	DestructLinter              bool         `json:"destruct_linter" yaml:"destruct_linter"`
	FloatDoubleLinter           bool         `json:"float_double_linter" yaml:"float_double_linter"`
	ForeignKeyLinter            bool         `json:"foreign_key_linter" yaml:"foreign_key_linter"`
	IDExistsLinter              bool         `json:"id_exists_linter" yaml:"id_exists_linter"`
	IDIsPrimaryLinter           bool         `json:"id_is_primary_linter" yaml:"id_is_primary_linter"`
	IDTypeLinter                bool         `json:"id_type_linter" yaml:"id_type_linter"`
	IndexLengthLinter           bool         `json:"index_length_linter" yaml:"index_length_linter"`
	IndexNameLinter             bool         `json:"index_name_linter" yaml:"index_name_linter"`
	KeywordsLinter              bool         `json:"keywords_linter" yaml:"keywords_linter"`
	NotNullLinter               bool         `json:"not_null_linter" yaml:"not_null_linter"`
	TableCommentLinter          bool         `json:"table_comment_linter" yaml:"table_comment_linter"`
	TableNameLinter             bool         `json:"table_name_linter" yaml:"table_name_linter"`
	UpdatedAtExistsLinter       bool         `json:"updated_at_exists_linter" yaml:"updated_at_exists_linter"`
	UpdatedAtTypeLinter         bool         `json:"updated_at_type_linter" yaml:"updated_at_type_linter"`
	UpdatedAtDefaultValueLinter bool         `json:"updated_at_default_value_linter" yaml:"updated_at_default_value_linter"`
	UpdatedAtOnUpdateLinter     bool         `json:"updated_at_on_update_linter" yaml:"updated_at_on_update_linter"`
	VarcharLengthLinter         bool         `json:"varchar_length_linter" yaml:"varchar_length_linter"`
	CompleteInsertLinter        bool         `json:"complete_insert_linter" yaml:"complete_insert_linter"`
	ManualTimeSetterLinter      bool         `json:"manual_time_setter_linter" yaml:"manual_time_setter_linter"`
	ExplicitCollationLinter     bool         `json:"explicit_collation_linter" yaml:"explicit_collation_linter"`
}

// ToJsonIndent marshals the Configuration to JSON []byte
func (c *Configuration) ToJsonIndent() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// ToYaml marshals the Configuration to YAML []byte
func (c *Configuration) ToYaml() ([]byte, error) {
	return yaml.Marshal(c)
}

// Rulers returns the rulers
func (c *Configuration) Rulers() (rulers []rules.Ruler, err error) {
	if c.AllowedDDLs == nil {
		return nil, errors.New("no DDL allowed, did you config it ?")
	}
	if c.AllowedDMLs == nil {
		return nil, errors.New("no DML allowed, did you config it ?")
	}

	var allowedDDLs uint64
	if c.AllowedDDLs.CreateDatabaseStmt {
		allowedDDLs |= linters.CreateDatabaseStmt
	}
	if c.AllowedDDLs.AlterDatabaseStmt {
		allowedDDLs |= linters.AlterDatabaseStmt
	}
	if c.AllowedDDLs.DropDatabaseStmt {
		allowedDDLs |= linters.DropDatabaseStmt
	}
	if c.AllowedDDLs.CreateTableStmt {
		allowedDDLs |= linters.CreateTableStmt
	}
	if c.AllowedDDLs.DropTableStmt {
		allowedDDLs |= linters.DropTableStmt
	}
	if c.AllowedDDLs.DropSequenceStmt {
		allowedDDLs |= linters.DropSequenceStmt
	}
	if c.AllowedDDLs.RenameTableStmt {
		allowedDDLs |= linters.RenameTableStmt
	}
	if c.AllowedDDLs.CreateViewStmt {
		allowedDDLs |= linters.CreateViewStmt
	}
	if c.AllowedDDLs.CreateSequenceStmt {
		allowedDDLs |= linters.CreateSequenceStmt
	}
	if c.AllowedDDLs.CreateIndexStmt {
		allowedDDLs |= linters.CreateIndexStmt
	}
	if c.AllowedDDLs.DropIndexStmt {
		allowedDDLs |= linters.DropIndexStmt
	}
	if c.AllowedDDLs.LockTablesStmt {
		allowedDDLs |= linters.LockTablesStmt
	}
	if c.AllowedDDLs.UnlockTablesStmt {
		allowedDDLs |= linters.UnlockTablesStmt
	}
	if c.AllowedDDLs.CleanupTableLockStmt {
		allowedDDLs |= linters.CleanupTableLockStmt
	}
	if c.AllowedDDLs.RepairTableStmt {
		allowedDDLs |= linters.RepairTableStmt
	}
	if c.AllowedDDLs.TruncateTableStmt {
		allowedDDLs |= linters.TruncateTableStmt
	}
	if c.AllowedDDLs.RecoverTableStmt {
		allowedDDLs |= linters.RecoverTableStmt
	}
	if c.AllowedDDLs.FlashBackTableStmt {
		allowedDDLs |= linters.FlashBackTableStmt
	}
	if c.AllowedDDLs.AlterTableOption {
		allowedDDLs |= linters.AlterTableOption
	}
	if c.AllowedDDLs.AlterTableAddColumns {
		allowedDDLs |= linters.AlterTableAddColumns
	}
	if c.AllowedDDLs.AlterTableAddConstraint {
		allowedDDLs |= linters.AlterTableAddConstraint
	}
	if c.AllowedDDLs.AlterTableDropColumn {
		allowedDDLs |= linters.AlterTableDropColumn
	}
	if c.AllowedDDLs.AlterTableDropPrimaryKey {
		allowedDDLs |= linters.AlterTableDropPrimaryKey
	}
	if c.AllowedDDLs.AlterTableDropIndex {
		allowedDDLs |= linters.AlterTableDropIndex
	}
	if c.AllowedDDLs.AlterTableDropForeignKey {
		allowedDDLs |= linters.AlterTableDropForeignKey
	}
	if c.AllowedDDLs.AlterTableModifyColumn {
		allowedDDLs |= linters.AlterTableModifyColumn
	}
	if c.AllowedDDLs.AlterTableChangeColumn {
		allowedDDLs |= linters.AlterTableChangeColumn
	}
	if c.AllowedDDLs.AlterTableRenameColumn {
		allowedDDLs |= linters.AlterTableRenameColumn
	}
	if c.AllowedDDLs.AlterTableRenameTable {
		allowedDDLs |= linters.AlterTableRenameTable
	}
	if c.AllowedDDLs.AlterTableAlterColumn {
		allowedDDLs |= linters.AlterTableAlterColumn
	}
	if c.AllowedDDLs.AlterTableLock {
		allowedDDLs |= linters.AlterTableLock
	}
	if c.AllowedDDLs.AlterTableAlgorithm {
		allowedDDLs |= linters.AlterTableAlgorithm
	}
	if c.AllowedDDLs.AlterTableRenameIndex {
		allowedDDLs |= linters.AlterTableRenameIndex
	}
	if c.AllowedDDLs.AlterTableForce {
		allowedDDLs |= linters.AlterTableForce
	}
	if c.AllowedDDLs.AlterTableAddPartitions {
		allowedDDLs |= linters.AlterTableAddPartitions
	}
	if c.AllowedDDLs.AlterTableCoalescePartitions {
		allowedDDLs |= linters.AlterTableCoalescePartitions
	}
	if c.AllowedDDLs.AlterTableDropPartition {
		allowedDDLs |= linters.AlterTableDropPartition
	}
	if c.AllowedDDLs.AlterTableTruncatePartition {
		allowedDDLs |= linters.AlterTableTruncatePartition
	}
	if c.AllowedDDLs.AlterTablePartition {
		allowedDDLs |= linters.AlterTablePartition
	}
	if c.AllowedDDLs.AlterTableEnableKeys {
		allowedDDLs |= linters.AlterTableEnableKeys
	}
	if c.AllowedDDLs.AlterTableDisableKeys {
		allowedDDLs |= linters.AlterTableDisableKeys
	}
	if c.AllowedDDLs.AlterTableRemovePartitioning {
		allowedDDLs |= linters.AlterTableRemovePartitioning
	}
	if c.AllowedDDLs.AlterTableWithValidation {
		allowedDDLs |= linters.AlterTableWithValidation
	}
	if c.AllowedDDLs.AlterTableWithoutValidation {
		allowedDDLs |= linters.AlterTableWithoutValidation
	}
	if c.AllowedDDLs.AlterTableSecondaryLoad {
		allowedDDLs |= linters.AlterTableSecondaryLoad
	}
	if c.AllowedDDLs.AlterTableSecondaryUnload {
		allowedDDLs |= linters.AlterTableSecondaryUnload
	}
	if c.AllowedDDLs.AlterTableRebuildPartition {
		allowedDDLs |= linters.AlterTableRebuildPartition
	}
	if c.AllowedDDLs.AlterTableReorganizePartition {
		allowedDDLs |= linters.AlterTableReorganizePartition
	}
	if c.AllowedDDLs.AlterTableCheckPartitions {
		allowedDDLs |= linters.AlterTableCheckPartitions
	}
	if c.AllowedDDLs.AlterTableExchangePartition {
		allowedDDLs |= linters.AlterTableExchangePartition
	}
	if c.AllowedDDLs.AlterTableOptimizePartition {
		allowedDDLs |= linters.AlterTableOptimizePartition
	}
	if c.AllowedDDLs.AlterTableRepairPartition {
		allowedDDLs |= linters.AlterTableRepairPartition
	}
	if c.AllowedDDLs.AlterTableImportPartitionTablespace {
		allowedDDLs |= linters.AlterTableImportPartitionTablespace
	}
	if c.AllowedDDLs.AlterTableDiscardPartitionTablespace {
		allowedDDLs |= linters.AlterTableDiscardPartitionTablespace
	}
	if c.AllowedDDLs.AlterTableAlterCheck {
		allowedDDLs |= linters.AlterTableAlterCheck
	}
	if c.AllowedDDLs.AlterTableDropCheck {
		allowedDDLs |= linters.AlterTableDropCheck
	}
	if c.AllowedDDLs.AlterTableImportTablespace {
		allowedDDLs |= linters.AlterTableImportTablespace
	}
	if c.AllowedDDLs.AlterTableDiscardTablespace {
		allowedDDLs |= linters.AlterTableDiscardTablespace
	}
	if c.AllowedDDLs.AlterTableIndexInvisible {
		allowedDDLs |= linters.AlterTableIndexInvisible
	}
	if c.AllowedDDLs.AlterTableOrderByColumns {
		allowedDDLs |= linters.AlterTableOrderByColumns
	}
	if c.AllowedDDLs.AlterTableSetTiFlashReplica {
		allowedDDLs |= linters.AlterTableSetTiFlashReplica
	}

	var allowedDMLs uint64
	if c.AllowedDMLs.SelectStmt {
		allowedDMLs |= linters.SelectStmt
	}
	if c.AllowedDMLs.UnionStmt {
		allowedDMLs |= linters.UnionStmt
	}
	if c.AllowedDMLs.LoadDataStmt {
		allowedDMLs |= linters.LoadDataStmt
	}
	if c.AllowedDMLs.InsertStmt {
		allowedDMLs |= linters.InsertStmt
	}
	if c.AllowedDMLs.DeleteStmt {
		allowedDMLs |= linters.DeleteStmt
	}
	if c.AllowedDMLs.UpdateStmt {
		allowedDMLs |= linters.UpdateStmt
	}
	if c.AllowedDMLs.ShowStmt {
		allowedDMLs |= linters.ShowStmt
	}
	if c.AllowedDMLs.SplitRegionStmt {
		allowedDMLs |= linters.SplitRegionStmt
	}

	rulers = append(rulers, linters.NewAllowedStmtLinter(allowedDDLs, allowedDMLs))
	if c.BooleanFieldLinter {
		rulers = append(rulers, linters.NewBooleanFieldLinter)
	}
	if c.CharsetLinter {
		rulers = append(rulers, linters.NewCharsetLinter)
	}
	if c.ColumnCommentLinter {
		rulers = append(rulers, linters.NewColumnCommentLinter)
	}
	if c.ColumnNameLinter {
		rulers = append(rulers, linters.NewColumnNameLinter)
	}
	if c.CreatedAtDefaultValueLinter {
		rulers = append(rulers, linters.NewCreatedAtDefaultValueLinter)
	}
	if c.CreatedAtExistsLinter {
		rulers = append(rulers, linters.NewCreatedAtExistsLinter)
	}
	if c.CreatedAtTypeLinter {
		rulers = append(rulers, linters.NewCreatedAtTypeLinter)
	}
	if c.DestructLinter {
		rulers = append(rulers, linters.NewDestructLinter)
	}
	if c.FloatDoubleLinter {
		rulers = append(rulers, linters.NewFloatDoubleLinter)
	}
	if c.ForeignKeyLinter {
		rulers = append(rulers, linters.NewForeignKeyLinter)
	}
	if c.IDExistsLinter {
		rulers = append(rulers, linters.NewIDExistsLinter)
	}
	if c.IDIsPrimaryLinter {
		rulers = append(rulers, linters.NewIDIsPrimaryLinter)
	}
	if c.IDTypeLinter {
		rulers = append(rulers, linters.NewIDTypeLinter)
	}
	if c.IndexLengthLinter {
		rulers = append(rulers, linters.NewIndexLengthLinter)
	}
	if c.IndexNameLinter {
		rulers = append(rulers, linters.NewIndexNameLinter)
	}
	if c.KeywordsLinter {
		rulers = append(rulers, linters.NewKeywordsLinter)
	}
	if c.NotNullLinter {
		rulers = append(rulers, linters.NewNotNullLinter)
	}
	if c.TableCommentLinter {
		rulers = append(rulers, linters.NewTableCommentLinter)
	}
	if c.TableNameLinter {
		rulers = append(rulers, linters.NewTableNameLinter)
	}
	if c.UpdatedAtExistsLinter {
		rulers = append(rulers, linters.NewUpdatedAtExistsLinter)
	}
	if c.UpdatedAtTypeLinter {
		rulers = append(rulers, linters.NewUpdatedAtTypeLinter)
	}
	if c.UpdatedAtDefaultValueLinter {
		rulers = append(rulers, linters.NewUpdatedAtDefaultValueLinter)
	}
	if c.UpdatedAtOnUpdateLinter {
		rulers = append(rulers, linters.NewUpdatedAtOnUpdateLinter)
	}
	if c.VarcharLengthLinter {
		rulers = append(rulers, linters.NewVarcharLengthLinter)
	}
	if c.CompleteInsertLinter {
		rulers = append(rulers, linters.NewCompleteInsertLinter)
	}
	if c.ManualTimeSetterLinter {
		rulers = append(rulers, linters.NewManualTimeSetterLinter)
	}
	if c.ExplicitCollationLinter {
		rulers = append(rulers, linters.NewExplicitCollationLinter)
	}

	return rulers, nil
}

type AllowedDDLs struct {
	CreateDatabaseStmt                   bool `json:"create_database_stmt" yaml:"create_database_stmt"`
	AlterDatabaseStmt                    bool `json:"alter_database_stmt" yaml:"alter_database_stmt"`
	DropDatabaseStmt                     bool `json:"drop_database_stmt" yaml:"drop_database_stmt"`
	CreateTableStmt                      bool `json:"create_table_stmt" yaml:"create_table_stmt"`
	DropTableStmt                        bool `json:"drop_table_stmt" yaml:"drop_table_stmt"`
	DropSequenceStmt                     bool `json:"drop_sequence_stmt" yaml:"drop_sequence_stmt"`
	RenameTableStmt                      bool `json:"rename_table_stmt" yaml:"rename_table_stmt"`
	CreateViewStmt                       bool `json:"create_view_stmt" yaml:"create_view_stmt"`
	CreateSequenceStmt                   bool `json:"create_sequence_stmt" yaml:"create_sequence_stmt"`
	CreateIndexStmt                      bool `json:"create_index_stmt" yaml:"create_index_stmt"`
	DropIndexStmt                        bool `json:"drop_index_stmt" yaml:"drop_index_stmt"`
	LockTablesStmt                       bool `json:"lock_tables_stmt" yaml:"lock_tables_stmt"`
	UnlockTablesStmt                     bool `json:"unlock_tables_stmt" yaml:"unlock_tables_stmt"`
	CleanupTableLockStmt                 bool `json:"cleanup_table_lock_stmt" yaml:"cleanup_table_lock_stmt"`
	RepairTableStmt                      bool `json:"repair_table_stmt" yaml:"repair_table_stmt"`
	TruncateTableStmt                    bool `json:"truncate_table_stmt" yaml:"truncate_table_stmt"`
	RecoverTableStmt                     bool `json:"recover_table_stmt" yaml:"recover_table_stmt"`
	FlashBackTableStmt                   bool `json:"flash_back_table_stmt" yaml:"flash_back_table_stmt"`
	AlterTableOption                     bool `json:"alter_table_option" yaml:"alter_table_option"`
	AlterTableAddColumns                 bool `json:"alter_table_add_columns" yaml:"alter_table_add_columns"`
	AlterTableAddConstraint              bool `json:"alter_table_add_constraint" yaml:"alter_table_add_constraint"`
	AlterTableDropColumn                 bool `json:"alter_table_drop_column" yaml:"alter_table_drop_column"`
	AlterTableDropPrimaryKey             bool `json:"alter_table_drop_primary_key" yaml:"alter_table_drop_primary_key"`
	AlterTableDropIndex                  bool `json:"alter_table_drop_index" yaml:"alter_table_drop_index"`
	AlterTableDropForeignKey             bool `json:"alter_table_drop_foreign_key" yaml:"alter_table_drop_foreign_key"`
	AlterTableModifyColumn               bool `json:"alter_table_modify_column" yaml:"alter_table_modify_column"`
	AlterTableChangeColumn               bool `json:"alter_table_change_column" yaml:"alter_table_change_column"`
	AlterTableRenameColumn               bool `json:"alter_table_rename_column" yaml:"alter_table_rename_column"`
	AlterTableRenameTable                bool `json:"alter_table_rename_table" yaml:"alter_table_rename_table"`
	AlterTableAlterColumn                bool `json:"alter_table_alter_column" yaml:"alter_table_alter_column"`
	AlterTableLock                       bool `json:"alter_table_lock" yaml:"alter_table_lock"`
	AlterTableAlgorithm                  bool `json:"alter_table_algorithm" yaml:"alter_table_algorithm"`
	AlterTableRenameIndex                bool `json:"alter_table_rename_index" yaml:"alter_table_rename_index"`
	AlterTableForce                      bool `json:"alter_table_force" yaml:"alter_table_force"`
	AlterTableAddPartitions              bool `json:"alter_table_add_partitions" yaml:"alter_table_add_partitions"`
	AlterTableCoalescePartitions         bool `json:"alter_table_coalesce_partitions" yaml:"alter_table_coalesce_partitions"`
	AlterTableDropPartition              bool `json:"alter_table_drop_partition" yaml:"alter_table_drop_partition"`
	AlterTableTruncatePartition          bool `json:"alter_table_truncate_partition" yaml:"alter_table_truncate_partition"`
	AlterTablePartition                  bool `json:"alter_table_partition" yaml:"alter_table_partition"`
	AlterTableEnableKeys                 bool `json:"alter_table_enable_keys" yaml:"alter_table_enable_keys"`
	AlterTableDisableKeys                bool `json:"alter_table_disable_keys" yaml:"alter_table_disable_keys"`
	AlterTableRemovePartitioning         bool `json:"alter_table_remove_partitioning" yaml:"alter_table_remove_partitioning"`
	AlterTableWithValidation             bool `json:"alter_table_with_validation" yaml:"alter_table_with_validation"`
	AlterTableWithoutValidation          bool `json:"alter_table_without_validation" yaml:"alter_table_without_validation"`
	AlterTableSecondaryLoad              bool `json:"alter_table_secondary_load" yaml:"alter_table_secondary_load"`
	AlterTableSecondaryUnload            bool `json:"alter_table_secondary_unload" yaml:"alter_table_secondary_unload"`
	AlterTableRebuildPartition           bool `json:"alter_table_rebuild_partition" yaml:"alter_table_rebuild_partition"`
	AlterTableReorganizePartition        bool `json:"alter_table_reorganize_partition" yaml:"alter_table_reorganize_partition"`
	AlterTableCheckPartitions            bool `json:"alter_table_check_partitions" yaml:"alter_table_check_partitions"`
	AlterTableExchangePartition          bool `json:"alter_table_exchange_partition" yaml:"alter_table_exchange_partition"`
	AlterTableOptimizePartition          bool `json:"alter_table_optimize_partition" yaml:"alter_table_optimize_partition"`
	AlterTableRepairPartition            bool `json:"alter_table_repair_partition" yaml:"alter_table_repair_partition"`
	AlterTableImportPartitionTablespace  bool `json:"alter_table_import_partition_tablespace" yaml:"alter_table_import_partition_tablespace"`
	AlterTableDiscardPartitionTablespace bool `json:"alter_table_discard_partition_tablespace" yaml:"alter_table_discard_partition_tablespace"`
	AlterTableAlterCheck                 bool `json:"alter_table_alter_check" yaml:"alter_table_alter_check"`
	AlterTableDropCheck                  bool `json:"alter_table_drop_check" yaml:"alter_table_drop_check"`
	AlterTableImportTablespace           bool `json:"alter_table_import_tablespace" yaml:"alter_table_import_tablespace"`
	AlterTableDiscardTablespace          bool `json:"alter_table_discard_tablespace" yaml:"alter_table_discard_tablespace"`
	AlterTableIndexInvisible             bool `json:"alter_table_index_invisible" yaml:"alter_table_index_invisible"`
	AlterTableOrderByColumns             bool `json:"alter_table_order_by_columns" yaml:"alter_table_order_by_columns"`
	AlterTableSetTiFlashReplica          bool `json:"alter_table_set_ti_flash_replica" yaml:"alter_table_set_ti_flash_replica"`
}

type AllowedDMLs struct {
	SelectStmt      bool `json:"select_stmt" yaml:"select_stmt"`
	UnionStmt       bool `json:"union_stmt" yaml:"union_stmt"`
	LoadDataStmt    bool `json:"load_data_stmt" yaml:"load_data_stmt"`
	InsertStmt      bool `json:"insert_stmt" yaml:"insert_stmt"`
	DeleteStmt      bool `json:"delete_stmt" yaml:"delete_stmt"`
	UpdateStmt      bool `json:"update_stmt" yaml:"update_stmt"`
	ShowStmt        bool `json:"show_stmt" yaml:"show_stmt"`
	SplitRegionStmt bool `json:"split_region_stmt" yaml:"split_region_stmt"`
}

// FromLocal reads the local file and returns the *Configuration
func FromLocal(filename string) (*Configuration, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FromData(data)
}

// FromData reads the data and returns the *Configuration
func FromData(data []byte) (*Configuration, error) {
	cfg := new(Configuration)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
