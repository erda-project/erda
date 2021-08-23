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

package linters

import (
	"fmt"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

const (
	CreateDatabaseStmt uint64 = 1 << iota
	AlterDatabaseStmt
	DropDatabaseStmt
	CreateTableStmt
	DropTableStmt
	DropSequenceStmt
	RenameTableStmt
	CreateViewStmt
	CreateSequenceStmt
	CreateIndexStmt
	DropIndexStmt
	LockTablesStmt
	UnlockTablesStmt
	CleanupTableLockStmt
	RepairTableStmt
	TruncateTableStmt
	RecoverTableStmt
	FlashBackTableStmt

	AlterTableOption
	AlterTableAddColumns
	AlterTableAddConstraint
	AlterTableDropColumn
	AlterTableDropPrimaryKey
	AlterTableDropIndex
	AlterTableDropForeignKey
	AlterTableModifyColumn
	AlterTableChangeColumn
	AlterTableRenameColumn
	AlterTableRenameTable
	AlterTableAlterColumn
	AlterTableLock
	AlterTableAlgorithm
	AlterTableRenameIndex
	AlterTableForce
	AlterTableAddPartitions
	AlterTableCoalescePartitions
	AlterTableDropPartition
	AlterTableTruncatePartition
	AlterTablePartition
	AlterTableEnableKeys
	AlterTableDisableKeys
	AlterTableRemovePartitioning
	AlterTableWithValidation
	AlterTableWithoutValidation
	AlterTableSecondaryLoad
	AlterTableSecondaryUnload
	AlterTableRebuildPartition
	AlterTableReorganizePartition
	AlterTableCheckPartitions
	AlterTableExchangePartition
	AlterTableOptimizePartition
	AlterTableRepairPartition
	AlterTableImportPartitionTablespace
	AlterTableDiscardPartitionTablespace
	AlterTableAlterCheck
	AlterTableDropCheck
	AlterTableImportTablespace
	AlterTableDiscardTablespace
	AlterTableIndexInvisible
	AlterTableOrderByColumns
	AlterTableSetTiFlashReplica
)

const (
	SelectStmt uint64 = 1 << iota
	UnionStmt
	LoadDataStmt
	InsertStmt
	DeleteStmt
	UpdateStmt
	ShowStmt
	SplitRegionStmt
)

const (
	AlterTableStmt = AlterTableOption | AlterTableAddColumns | AlterTableAddConstraint |
		AlterTableDropColumn | AlterTableDropPrimaryKey | AlterTableDropIndex |
		AlterTableDropForeignKey | AlterTableModifyColumn |
		AlterTableChangeColumn | AlterTableRenameColumn | AlterTableRenameTable | AlterTableAlterColumn |
		AlterTableLock | AlterTableAlgorithm | AlterTableRenameIndex |
		AlterTableForce | AlterTableAddPartitions | AlterTableCoalescePartitions |
		AlterTableDropPartition | AlterTableTruncatePartition | AlterTablePartition |
		AlterTableEnableKeys | AlterTableDisableKeys | AlterTableRemovePartitioning |
		AlterTableWithValidation | AlterTableWithoutValidation | AlterTableSecondaryLoad |
		AlterTableSecondaryUnload | AlterTableRebuildPartition | AlterTableReorganizePartition |
		AlterTableCheckPartitions | AlterTableExchangePartition | AlterTableOptimizePartition |
		AlterTableRepairPartition | AlterTableImportPartitionTablespace | AlterTableDiscardPartitionTablespace |
		AlterTableAlterCheck | AlterTableDropCheck | AlterTableImportTablespace |
		AlterTableDiscardTablespace | AlterTableIndexInvisible | AlterTableOrderByColumns |
		AlterTableSetTiFlashReplica

	DDLStmt = CreateDatabaseStmt | AlterDatabaseStmt | DropDatabaseStmt |
		CreateTableStmt | DropTableStmt | DropSequenceStmt |
		RenameTableStmt | CreateViewStmt | CreateSequenceStmt |
		CreateIndexStmt | DropIndexStmt | LockTablesStmt |
		UnlockTablesStmt | CleanupTableLockStmt | RepairTableStmt |
		TruncateTableStmt | RecoverTableStmt | FlashBackTableStmt |
		AlterTableStmt

	DMLStmt = SelectStmt | UnionStmt | LoadDataStmt |
		InsertStmt | DeleteStmt | UpdateStmt |
		ShowStmt | SplitRegionStmt
)

type AllowedStmt struct {
	baseLinter

	allowedDDL uint64
	allowedDML uint64
}

func NewAllowedStmtLinter(allowedDDLs, allowedDMLs uint64) rules.Ruler {
	return func(script script.Script) rules.Rule {
		return &AllowedStmt{
			baseLinter: baseLinter{
				s:    script,
				err:  nil,
				text: "",
			},
			allowedDDL: allowedDDLs,
			allowedDML: allowedDMLs,
		}
	}
}

func (l *AllowedStmt) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case *ast.CreateDatabaseStmt:
		if l.allowedDDL&CreateDatabaseStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CreateDatabaseStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.AlterDatabaseStmt:
		if l.allowedDDL&AlterDatabaseStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed AlterDatabaseStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.DropDatabaseStmt:
		if l.allowedDDL&DropDatabaseStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed DropDatabaseStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.CreateTableStmt:
		if l.allowedDDL&CreateTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CreateTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.DropTableStmt:
		if l.allowedDDL&DropTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed DropTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.DropSequenceStmt:
		if l.allowedDDL&DropSequenceStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed DropSequenceStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.RenameTableStmt:
		if l.allowedDDL&RenameTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed RenameTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.CreateViewStmt:
		if l.allowedDDL&CreateViewStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CreateViewStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.CreateSequenceStmt:
		if l.allowedDDL&CreateSequenceStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CreateSequenceStmt", func([]byte) bool {
				return true
			})
			return in, true
		}

	case *ast.CreateIndexStmt:
		if l.allowedDDL&CreateIndexStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CreateIndexStmt", func([]byte) bool {
				return true
			})
			return in, true
		}

	case *ast.DropIndexStmt:
		if l.allowedDDL&DropIndexStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed DropIndexStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.LockTablesStmt:
		if l.allowedDDL&LockTablesStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed LockTablesStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.UnlockTablesStmt:
		if l.allowedDDL&LockTablesStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed LockTablesStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.CleanupTableLockStmt:
		if l.allowedDDL&CleanupTableLockStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed CleanupTableLockStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.RepairTableStmt:
		if l.allowedDDL&RepairTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed RepairTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.TruncateTableStmt:
		if l.allowedDDL&TruncateTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed TruncateTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.RecoverTableStmt:
		if l.allowedDDL&RecoverTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed RecoverTableStmt", func([]byte) bool {
				return true
			})
		}

	case *ast.FlashBackTableStmt:
		if l.allowedDDL&FlashBackTableStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed FlashBackTableStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.AlterTableStmt:
		for _, spec := range in.(*ast.AlterTableStmt).Specs {
			switch spec.Tp {
			case ast.AlterTableOption:
				if l.allowedDDL&AlterTableOption == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableOption", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAddColumns:
				if l.allowedDDL&AlterTableAddColumns == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAddColumns", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAddConstraint:
				if l.allowedDDL&AlterTableAddConstraint == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAddConstraint", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropColumn:
				if l.allowedDDL&AlterTableDropColumn == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropColumn", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropPrimaryKey:
				if l.allowedDDL&AlterTableDropPrimaryKey == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropPrimaryKey", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropIndex:
				if l.allowedDDL&AlterTableDropIndex == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropIndex", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropForeignKey:
				if l.allowedDDL&AlterTableDropForeignKey == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropForeignKey", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableModifyColumn:
				if l.allowedDDL&AlterTableModifyColumn == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableModifyColumn", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableChangeColumn:
				if l.allowedDDL&AlterTableChangeColumn == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableChangeColumn", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRenameColumn:
				if l.allowedDDL&AlterTableRenameColumn == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRenameColumn", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRenameTable:
				if l.allowedDDL&AlterTableRenameTable == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRenameTable", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAlterColumn:
				if l.allowedDDL&AlterTableAlterColumn == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAlterColumn", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableLock:
				if l.allowedDDL&AlterTableLock == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableLock", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAlgorithm:
				if l.allowedDDL&AlterTableAlgorithm == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAlgorithm", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRenameIndex:
				if l.allowedDDL&AlterTableRenameIndex == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRenameIndex", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableForce:
				if l.allowedDDL&AlterTableForce == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableForce", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAddPartitions:
				if l.allowedDDL&AlterTableAddPartitions == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAddPartitions", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableCoalescePartitions:
				if l.allowedDDL&AlterTableCoalescePartitions == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableCoalescePartitions", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropPartition:
				if l.allowedDDL&AlterTableDropPartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropPartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableTruncatePartition:
				if l.allowedDDL&AlterTableTruncatePartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableTruncatePartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTablePartition:
				if l.allowedDDL&AlterTablePartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTablePartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableEnableKeys:
				if l.allowedDDL&AlterTableEnableKeys == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableEnableKeys", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDisableKeys:
				if l.allowedDDL&AlterTableDisableKeys == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDisableKeys", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRemovePartitioning:
				if l.allowedDDL&AlterTableRemovePartitioning == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRemovePartitioning", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableWithValidation:
				if l.allowedDDL&AlterTableWithValidation == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableWithValidation", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableWithoutValidation:
				if l.allowedDDL&AlterTableWithoutValidation == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableWithoutValidation", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableSecondaryLoad:
				if l.allowedDDL&AlterTableSecondaryLoad == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableSecondaryLoad", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableSecondaryUnload:
				if l.allowedDDL&AlterTableSecondaryUnload == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableSecondaryUnload", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRebuildPartition:
				if l.allowedDDL&AlterTableRebuildPartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRebuildPartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableReorganizePartition:
				if l.allowedDDL&AlterTableReorganizePartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableReorganizePartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableCheckPartitions:
				if l.allowedDDL&AlterTableCheckPartitions == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableCheckPartitions", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableExchangePartition:
				if l.allowedDDL&AlterTableExchangePartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableExchangePartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableOptimizePartition:
				if l.allowedDDL&AlterTableOptimizePartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableOptimizePartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableRepairPartition:
				if l.allowedDDL&AlterTableRepairPartition == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableRepairPartition", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableImportPartitionTablespace:
				if l.allowedDDL&AlterTableImportPartitionTablespace == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableImportPartitionTablespace", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDiscardPartitionTablespace:
				if l.allowedDDL&AlterTableDiscardPartitionTablespace == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDiscardPartitionTablespace", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableAlterCheck:
				if l.allowedDDL&AlterTableAlterCheck == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableAlterCheck", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDropCheck:
				if l.allowedDDL&AlterTableDropCheck == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDropCheck", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableImportTablespace:
				if l.allowedDDL&AlterTableImportTablespace == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableImportTablespace", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableDiscardTablespace:
				if l.allowedDDL&AlterTableDiscardTablespace == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableDiscardTablespace", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableIndexInvisible:
				if l.allowedDDL&AlterTableIndexInvisible == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableIndexInvisible", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableOrderByColumns:
				if l.allowedDDL&AlterTableOrderByColumns == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableOrderByColumns", func([]byte) bool {
						return true
					})
				}
			case ast.AlterTableSetTiFlashReplica:
				if l.allowedDDL&AlterTableSetTiFlashReplica == 0 {
					l.err = linterror.New(l.s, l.text, "not allowed AlterTableSetTiFlashReplica", func([]byte) bool {
						return true
					})
				}
			}
		}

	case *ast.SelectStmt:
		if l.allowedDML&SelectStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed SelectStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.UnionStmt:
		if l.allowedDML&UnionStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed UnionStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.LoadDataStmt:
		if l.allowedDML&LoadDataStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed LoadDataStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.InsertStmt:
		if l.allowedDML&InsertStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed InsertStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.DeleteStmt:
		if l.allowedDML&DeleteStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed DeleteStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.UpdateStmt:
		if l.allowedDML&UpdateStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed UpdateStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.ShowStmt:
		if l.allowedDML&ShowStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed ShowStmt", func([]byte) bool {
				return true
			})
		}
	case *ast.SplitRegionStmt:
		if l.allowedDML&SplitRegionStmt == 0 {
			l.err = linterror.New(l.s, l.text, "not allowed SplitRegionStmt", func([]byte) bool {
				return true
			})
		}

	default:
		l.err = linterror.New(l.s, l.text, fmt.Sprintf("not allowed stmt %T", in), func([]byte) bool {
			return true
		})
	}

	return in, true
}

func (l *AllowedStmt) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *AllowedStmt) Error() error {
	return l.err
}
