// Copyright (c) $date.year Terminus, Inc.
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

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/erda-project/erda/pkg/sqllint"
	"github.com/erda-project/erda/pkg/sqllint/linters"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	defaultAllowedDDLs = linters.CreateTableStmt | linters.CreateIndexStmt | linters.DropIndexStmt |
		linters.AlterTableOption | linters.AlterTableAddColumns | linters.AlterTableAddConstraint |
		linters.AlterTableDropIndex | linters.AlterTableModifyColumn | linters.AlterTableChangeColumn |
		linters.AlterTableAlterColumn | linters.AlterTableRenameIndex
	defaultAllowedDMLs = linters.SelectStmt | linters.UnionStmt | linters.InsertStmt |
		linters.UpdateStmt | linters.ShowStmt
)

const (
	baseScriptLabel  = "# MIGRATION_BASE"
	baseScriptLabel2 = "-- MIGRATION_BASE"
	baseScriptLabel3 = "/* MIGRATION_BASE */"
)

var defaultRulers = []rules.Ruler{
	linters.NewAllowedStmtLinter(defaultAllowedDDLs, defaultAllowedDMLs),
	linters.NewBooleanFieldLinter,
	linters.NewCharsetLinter,
	linters.NewColumnCommentLinter,
	linters.NewColumnNameLinter,

	linters.NewCreatedAtDefaultValueLinter,
	linters.NewCreatedAtExistsLinter,
	linters.NewCreatedAtTypeLinter,

	linters.NewDestructLinter,
	linters.NewFloatDoubleLinter,
	linters.NewForeignKeyLinter,

	linters.NewIDExistsLinter,
	linters.NewIDIsPrimaryLinter,
	linters.NewIDTypeLinter,

	linters.NewIndexLengthLinter,
	linters.NewIndexNameLinter,

	linters.NewKeywordsLinter,
	linters.NewNotNullLinter,
	linters.NewTableCommentLinter,
	linters.NewTableNameLinter,

	linters.NewUpdatedAtExistsLinter,
	linters.NewUpdatedAtTypeLinter,
	linters.NewUpdatedAtDefaultValueLinter,
	linters.NewUpdatedAtOnUpdateLinter,

	linters.NewVarcharLengthLinter,
}

var MigrationLint = command.Command{
	ParentName:     "",
	Name:           "miglint",
	ShortHelp:      "Erda MySQL Migration lint",
	LongHelp:       "Erda MySQL Migration lint",
	Example:        "erda-cli miglint --input=. config=default.yaml --detail",
	Hidden:         false,
	DontHideCursor: false,
	Args:           nil,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "f",
			Name:         "filename",
			Doc:          "[optional] the file or directory for linting",
			DefaultValue: ".",
		},
		command.StringFlag{
			Short:        "c",
			Name:         "config",
			Doc:          "[optional] the lint config file",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "d",
			Name:         "detail",
			Doc:          "[optional] print details of lint result",
			DefaultValue: true,
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          "[optional] result output file name",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "i",
			Name:         "ignoreBase",
			Doc:          "ignore script which marked baseline",
			DefaultValue: true,
		},
	},
	Run: RunMigrationLint,
}

func RunMigrationLint(ctx *command.Context, input, config string, detail bool, output string, ignoreBase bool) error {
	log.Printf("Erda MySQL Lint the input file or directory: %s", input)
	files := new(walk).walk(input, ".sql").filenames()

	rulers, err := rulers(config)
	if err != nil {
		return errors.Wrap(err, "failed to get lint rulers")
	}
	linter := sqllint.New(rulers...)

	for _, filename := range files {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "failed to open file, filename: %s", filename)
		}
		if ignoreBase && isBaseScript(data) {
			continue
		}
		if err := linter.Input(data, filename); err != nil {
			return errors.Wrapf(err, "failed to run Erda MySQL Lint on the SQL script, filename: %s", filename)
		}
	}

	if len(linter.Errors()) == 0 {
		log.Println("Erda MySQL Lint pass")
	}

	var out = log.Writer()
	if output != "" {

		f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("failed to OpenFile, filename: %s", output)
		} else {
			defer f.Close()
			out = f
		}
	}

	if detail {
		for src, errs := range linter.Errors() {
			if _, err := fmt.Fprintln(out, src); err != nil {
				return errors.Wrapf(err, "failed to print lint error")
			}
			for _, e := range errs {
				if _, err := fmt.Fprintln(out, e); err != nil {
					return errors.Wrapf(err, "failed to print lint error")
				}
			}
		}
		return nil
	}

	if _, err := fmt.Fprintln(out, linter.Report()); err != nil {
		return errors.Wrapf(err, "failed to print lint report")
	}

	return nil
}

type configuration struct {
	AllowedDDLs                 *allowedDDLs `json:"allowed_ddl" yaml:"allowed_ddl"`
	AllowedDMLs                 *allowedDMLs `json:"allowed_dml" yaml:"allowed_dml"`
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
}

type allowedDDLs struct {
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

type allowedDMLs struct {
	SelectStmt      bool `json:"select_stmt" yaml:"select_stmt"`
	UnionStmt       bool `json:"union_stmt" yaml:"union_stmt"`
	LoadDataStmt    bool `json:"load_data_stmt" yaml:"load_data_stmt"`
	InsertStmt      bool `json:"insert_stmt" yaml:"insert_stmt"`
	DeleteStmt      bool `json:"delete_stmt" yaml:"delete_stmt"`
	UpdateStmt      bool `json:"update_stmt" yaml:"update_stmt"`
	ShowStmt        bool `json:"show_stmt" yaml:"show_stmt"`
	SplitRegionStmt bool `json:"split_region_stmt" yaml:"split_region_stmt"`
}

func rulers(config string) (results []rules.Ruler, err error) {
	if config == "" {
		return defaultRulers, nil
	}

	data, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}

	var cfg configuration
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.AllowedDDLs == nil {
		return nil, errors.New("no DDL allowed, did you config it ?")
	}
	if cfg.AllowedDMLs == nil {
		return nil, errors.New("no DML allowed, did you config it ?")
	}

	var allowedDDLs uint64
	if cfg.AllowedDDLs.CreateDatabaseStmt {
		allowedDDLs |= linters.CreateDatabaseStmt
	}
	if cfg.AllowedDDLs.AlterDatabaseStmt {
		allowedDDLs |= linters.AlterDatabaseStmt
	}
	if cfg.AllowedDDLs.DropDatabaseStmt {
		allowedDDLs |= linters.DropDatabaseStmt
	}
	if cfg.AllowedDDLs.CreateTableStmt {
		allowedDDLs |= linters.CreateTableStmt
	}
	if cfg.AllowedDDLs.DropTableStmt {
		allowedDDLs |= linters.DropTableStmt
	}
	if cfg.AllowedDDLs.DropSequenceStmt {
		allowedDDLs |= linters.DropSequenceStmt
	}
	if cfg.AllowedDDLs.RenameTableStmt {
		allowedDDLs |= linters.RenameTableStmt
	}
	if cfg.AllowedDDLs.CreateViewStmt {
		allowedDDLs |= linters.CreateViewStmt
	}
	if cfg.AllowedDDLs.CreateSequenceStmt {
		allowedDDLs |= linters.CreateSequenceStmt
	}
	if cfg.AllowedDDLs.CreateIndexStmt {
		allowedDDLs |= linters.CreateIndexStmt
	}
	if cfg.AllowedDDLs.DropIndexStmt {
		allowedDDLs |= linters.DropIndexStmt
	}
	if cfg.AllowedDDLs.LockTablesStmt {
		allowedDDLs |= linters.LockTablesStmt
	}
	if cfg.AllowedDDLs.UnlockTablesStmt {
		allowedDDLs |= linters.UnlockTablesStmt
	}
	if cfg.AllowedDDLs.CleanupTableLockStmt {
		allowedDDLs |= linters.CleanupTableLockStmt
	}
	if cfg.AllowedDDLs.RepairTableStmt {
		allowedDDLs |= linters.RepairTableStmt
	}
	if cfg.AllowedDDLs.TruncateTableStmt {
		allowedDDLs |= linters.TruncateTableStmt
	}
	if cfg.AllowedDDLs.RecoverTableStmt {
		allowedDDLs |= linters.RecoverTableStmt
	}
	if cfg.AllowedDDLs.FlashBackTableStmt {
		allowedDDLs |= linters.FlashBackTableStmt
	}
	if cfg.AllowedDDLs.AlterTableOption {
		allowedDDLs |= linters.AlterTableOption
	}
	if cfg.AllowedDDLs.AlterTableAddColumns {
		allowedDDLs |= linters.AlterTableAddColumns
	}
	if cfg.AllowedDDLs.AlterTableAddConstraint {
		allowedDDLs |= linters.AlterTableAddConstraint
	}
	if cfg.AllowedDDLs.AlterTableDropColumn {
		allowedDDLs |= linters.AlterTableDropColumn
	}
	if cfg.AllowedDDLs.AlterTableDropPrimaryKey {
		allowedDDLs |= linters.AlterTableDropPrimaryKey
	}
	if cfg.AllowedDDLs.AlterTableDropIndex {
		allowedDDLs |= linters.AlterTableDropIndex
	}
	if cfg.AllowedDDLs.AlterTableDropForeignKey {
		allowedDDLs |= linters.AlterTableDropForeignKey
	}
	if cfg.AllowedDDLs.AlterTableModifyColumn {
		allowedDDLs |= linters.AlterTableModifyColumn
	}
	if cfg.AllowedDDLs.AlterTableChangeColumn {
		allowedDDLs |= linters.AlterTableChangeColumn
	}
	if cfg.AllowedDDLs.AlterTableRenameColumn {
		allowedDDLs |= linters.AlterTableRenameColumn
	}
	if cfg.AllowedDDLs.AlterTableRenameTable {
		allowedDDLs |= linters.AlterTableRenameTable
	}
	if cfg.AllowedDDLs.AlterTableAlterColumn {
		allowedDDLs |= linters.AlterTableAlterColumn
	}
	if cfg.AllowedDDLs.AlterTableLock {
		allowedDDLs |= linters.AlterTableLock
	}
	if cfg.AllowedDDLs.AlterTableAlgorithm {
		allowedDDLs |= linters.AlterTableAlgorithm
	}
	if cfg.AllowedDDLs.AlterTableRenameIndex {
		allowedDDLs |= linters.AlterTableRenameIndex
	}
	if cfg.AllowedDDLs.AlterTableForce {
		allowedDDLs |= linters.AlterTableForce
	}
	if cfg.AllowedDDLs.AlterTableAddPartitions {
		allowedDDLs |= linters.AlterTableAddPartitions
	}
	if cfg.AllowedDDLs.AlterTableCoalescePartitions {
		allowedDDLs |= linters.AlterTableCoalescePartitions
	}
	if cfg.AllowedDDLs.AlterTableDropPartition {
		allowedDDLs |= linters.AlterTableDropPartition
	}
	if cfg.AllowedDDLs.AlterTableTruncatePartition {
		allowedDDLs |= linters.AlterTableTruncatePartition
	}
	if cfg.AllowedDDLs.AlterTablePartition {
		allowedDDLs |= linters.AlterTablePartition
	}
	if cfg.AllowedDDLs.AlterTableEnableKeys {
		allowedDDLs |= linters.AlterTableEnableKeys
	}
	if cfg.AllowedDDLs.AlterTableDisableKeys {
		allowedDDLs |= linters.AlterTableDisableKeys
	}
	if cfg.AllowedDDLs.AlterTableRemovePartitioning {
		allowedDDLs |= linters.AlterTableRemovePartitioning
	}
	if cfg.AllowedDDLs.AlterTableWithValidation {
		allowedDDLs |= linters.AlterTableWithValidation
	}
	if cfg.AllowedDDLs.AlterTableWithoutValidation {
		allowedDDLs |= linters.AlterTableWithoutValidation
	}
	if cfg.AllowedDDLs.AlterTableSecondaryLoad {
		allowedDDLs |= linters.AlterTableSecondaryLoad
	}
	if cfg.AllowedDDLs.AlterTableSecondaryUnload {
		allowedDDLs |= linters.AlterTableSecondaryUnload
	}
	if cfg.AllowedDDLs.AlterTableRebuildPartition {
		allowedDDLs |= linters.AlterTableRebuildPartition
	}
	if cfg.AllowedDDLs.AlterTableReorganizePartition {
		allowedDDLs |= linters.AlterTableReorganizePartition
	}
	if cfg.AllowedDDLs.AlterTableCheckPartitions {
		allowedDDLs |= linters.AlterTableCheckPartitions
	}
	if cfg.AllowedDDLs.AlterTableExchangePartition {
		allowedDDLs |= linters.AlterTableExchangePartition
	}
	if cfg.AllowedDDLs.AlterTableOptimizePartition {
		allowedDDLs |= linters.AlterTableOptimizePartition
	}
	if cfg.AllowedDDLs.AlterTableRepairPartition {
		allowedDDLs |= linters.AlterTableRepairPartition
	}
	if cfg.AllowedDDLs.AlterTableImportPartitionTablespace {
		allowedDDLs |= linters.AlterTableImportPartitionTablespace
	}
	if cfg.AllowedDDLs.AlterTableDiscardPartitionTablespace {
		allowedDDLs |= linters.AlterTableDiscardPartitionTablespace
	}
	if cfg.AllowedDDLs.AlterTableAlterCheck {
		allowedDDLs |= linters.AlterTableAlterCheck
	}
	if cfg.AllowedDDLs.AlterTableDropCheck {
		allowedDDLs |= linters.AlterTableDropCheck
	}
	if cfg.AllowedDDLs.AlterTableImportTablespace {
		allowedDDLs |= linters.AlterTableImportTablespace
	}
	if cfg.AllowedDDLs.AlterTableDiscardTablespace {
		allowedDDLs |= linters.AlterTableDiscardTablespace
	}
	if cfg.AllowedDDLs.AlterTableIndexInvisible {
		allowedDDLs |= linters.AlterTableIndexInvisible
	}
	if cfg.AllowedDDLs.AlterTableOrderByColumns {
		allowedDDLs |= linters.AlterTableOrderByColumns
	}
	if cfg.AllowedDDLs.AlterTableSetTiFlashReplica {
		allowedDDLs |= linters.AlterTableSetTiFlashReplica
	}

	var allowedDMLs uint64
	if cfg.AllowedDMLs.SelectStmt {
		allowedDMLs |= linters.SelectStmt
	}
	if cfg.AllowedDMLs.UnionStmt {
		allowedDMLs |= linters.UnionStmt
	}
	if cfg.AllowedDMLs.LoadDataStmt {
		allowedDMLs |= linters.LoadDataStmt
	}
	if cfg.AllowedDMLs.InsertStmt {
		allowedDMLs |= linters.InsertStmt
	}
	if cfg.AllowedDMLs.DeleteStmt {
		allowedDMLs |= linters.DeleteStmt
	}
	if cfg.AllowedDMLs.UpdateStmt {
		allowedDMLs |= linters.UpdateStmt
	}
	if cfg.AllowedDMLs.ShowStmt {
		allowedDMLs |= linters.ShowStmt
	}
	if cfg.AllowedDMLs.SplitRegionStmt {
		allowedDMLs |= linters.SplitRegionStmt
	}

	results = append(results, linters.NewAllowedStmtLinter(allowedDDLs, allowedDMLs))
	if cfg.BooleanFieldLinter {
		results = append(results, linters.NewBooleanFieldLinter)
	}
	if cfg.CharsetLinter {
		results = append(results, linters.NewCharsetLinter)
	}
	if cfg.ColumnCommentLinter {
		results = append(results, linters.NewColumnCommentLinter)
	}
	if cfg.ColumnNameLinter {
		results = append(results, linters.NewColumnNameLinter)
	}
	if cfg.CreatedAtDefaultValueLinter {
		results = append(results, linters.NewCreatedAtDefaultValueLinter)
	}
	if cfg.CreatedAtExistsLinter {
		results = append(results, linters.NewCreatedAtExistsLinter)
	}
	if cfg.CreatedAtTypeLinter {
		results = append(results, linters.NewCreatedAtTypeLinter)
	}
	if cfg.DestructLinter {
		results = append(results, linters.NewDestructLinter)
	}
	if cfg.FloatDoubleLinter {
		results = append(results, linters.NewFloatDoubleLinter)
	}
	if cfg.ForeignKeyLinter {
		results = append(results, linters.NewForeignKeyLinter)
	}
	if cfg.IDExistsLinter {
		results = append(results, linters.NewIDExistsLinter)
	}
	if cfg.IDIsPrimaryLinter {
		results = append(results, linters.NewIDIsPrimaryLinter)
	}
	if cfg.IDTypeLinter {
		results = append(results, linters.NewIDTypeLinter)
	}
	if cfg.IndexLengthLinter {
		results = append(results, linters.NewIndexLengthLinter)
	}
	if cfg.IndexNameLinter {
		results = append(results, linters.NewIndexNameLinter)
	}
	if cfg.KeywordsLinter {
		results = append(results, linters.NewKeywordsLinter)
	}
	if cfg.NotNullLinter {
		results = append(results, linters.NewNotNullLinter)
	}
	if cfg.TableCommentLinter {
		results = append(results, linters.NewTableCommentLinter)
	}
	if cfg.TableNameLinter {
		results = append(results, linters.NewTableNameLinter)
	}
	if cfg.UpdatedAtExistsLinter {
		results = append(results, linters.NewUpdatedAtExistsLinter)
	}
	if cfg.UpdatedAtTypeLinter {
		results = append(results, linters.NewUpdatedAtTypeLinter)
	}
	if cfg.UpdatedAtDefaultValueLinter {
		results = append(results, linters.NewUpdatedAtDefaultValueLinter)
	}
	if cfg.UpdatedAtOnUpdateLinter {
		results = append(results, linters.NewUpdatedAtOnUpdateLinter)
	}
	if cfg.VarcharLengthLinter {
		results = append(results, linters.NewVarcharLengthLinter)
	}

	return results, nil
}

type walk struct {
	files []string
}

func (w *walk) filenames() []string {
	return w.files
}

func (w *walk) walk(input, suffix string) *walk {
	infos, err := ioutil.ReadDir(input)
	if err != nil {
		w.files = append(w.files, input)
		return w
	}

	for _, info := range infos {
		if info.IsDir() {
			w.walk(filepath.Join(input, info.Name()), suffix)
			continue
		}
		if strings.ToLower(path.Ext(info.Name())) == strings.ToLower(suffix) {
			file := filepath.Join(input, info.Name())
			w.files = append(w.files, file)
		}
	}

	return w
}

func isBaseScript(data []byte) bool {
	return bytes.HasPrefix(data, []byte(baseScriptLabel)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel2)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel3))
}
