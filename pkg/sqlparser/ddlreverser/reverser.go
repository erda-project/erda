package ddlreverser

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// ReverseDDLWithSnapshot reverses a DDL.
// e.g.
// 	CreateTableStmt ==reverse-to==> DropTableStmt
//	RenameTableStmt a to b ==reverse-to==> RenameTableStmt b to a
// 	AlterTableAddColumns ==reverse-to==> AlterTableDropColumn
// 	AlterTableOption to new ==reverse-to==> AlterTableOption to old
// 	... ....
//
// How to do this ?
// First, to snapshot the table definition from database, then generate the reversed DDL.
//
// tx *gorm.DB is used to connect to database for getting table's snapshot;
// ddl ast.DDLNode is is the stmt node for revering.
//
// reversing is the reversed DDL text;
// ok is whether reversing is success, if not ok, reversing is "";
// err is whether an error has occurred, if err is not nil, ok is false, reversing is "".
//
// If the ddl is not allowed (e.g. DropDatabaseStmt, DropTableStmt, AlterTableDropColumnSpec),
// the function will not process it, and returned err is nil, ok is false, reversing is "".
func ReverseDDLWithSnapshot(tx *gorm.DB, ddl ast.DDLNode) (reversing string, ok bool, err error) {
	switch ddl.(type) {

	// CREATE [TEMPORARY] TABLE [IF NOT EXISTS] tbl_name
	//    (create_definition,...)
	//    [table_options]
	//    [partition_options]
	case *ast.CreateTableStmt:
		tableName := ddl.(*ast.CreateTableStmt).Table.Name.String()
		return "DROP TABLE IF EXISTS " + tableName + ";\n", true, nil

	// RENAME TABLE
	//    tbl_name TO new_tbl_name
	//    [, tbl_name2 TO new_tbl_name2] ...
	case *ast.RenameTableStmt:
		stmt := ddl.(*ast.RenameTableStmt)
		var buf = bytes.NewBufferString("RENAME TABLE ")
		for i, t2t := range stmt.TableToTables {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(t2t.NewTable.Name.String())
			buf.WriteString(" TO ")
			buf.WriteString(t2t.OldTable.Name.String())
		}
		buf.WriteString(";\n")
		return buf.String(), true, nil

	// not allowed ddl
	case *ast.DropDatabaseStmt, *ast.DropTableStmt:
		return "", false, nil

	case *ast.DropIndexStmt:
		stmt := ddl.(*ast.DropIndexStmt)

		tableName := stmt.Table.Name.String()
		createTableSQL, err := ShowCreateTable(tx, tableName)
		if err != nil {
			return "", false, errors.Wrapf(err, "failed to ShowCreateTable %s", tableName)
		}
		createStmt, err := parser.New().ParseOneStmt(createTableSQL, "", "")
		if err != nil {
			return "", false, err
		}
		snap := createStmt.(*ast.CreateTableStmt)
		return ReverseDropIndexStmtWithCompares(snap, stmt)

	// AlterTableStmt
	// https://dev.mysql.com/doc/refman/8.0/en/alter-table.html
	case *ast.AlterTableStmt:
		alterTableStmt := ddl.(*ast.AlterTableStmt)
		tableName := alterTableStmt.Table.Name.String()

		// snapshot table's definition
		creatTableSQL, err := ShowCreateTable(tx, tableName)
		if err != nil {
			return "", false, errors.Wrapf(err, "failed to ShowCreateTable %s", tableName)
		}
		createStmt, err := parser.New().ParseOneStmt(creatTableSQL, "", "")
		if err != nil {
			return "", false, err
		}
		snap := createStmt.(*ast.CreateTableStmt)

		return ReverseAlterWithCompares(snap, alterTableStmt)
	}
	return "", false, nil
}

// ReverseCreateTableStmtsToDropTableStmts reverses DDLs without snapshot.
// Generally, this function is used to process the baseline,
// because when processing the baseline, it only needs to Drop all the newly created tables.
func ReverseCreateTableStmtsToDropTableStmts(ddlNodes interface{ DDLNodes() []ast.DDLNode }) string {
	var buf = bytes.NewBuffer(nil)
	for _, ddl := range ddlNodes.DDLNodes() {
		if create, ok := ddl.(*ast.CreateTableStmt); ok {
			tableName := create.Table.Name.String()
			buf.WriteString("DROP TABLE IF EXISTS " + tableName + ";\n")
		}
	}

	return buf.String()
}

// ShowCreateTable gets create table stmt by executing SHOW CREATE TABLE .
func ShowCreateTable(tx *gorm.DB, tableName string) (create string, err error) {
	if tableName == "" {
		return "", errors.New("the tableName can not be \"\"")
	}
	tx = tx.Raw("SHOW CREATE TABLE " + tableName)
	if err = tx.Error; err != nil {
		return "", err
	}
	if err = tx.Row().Scan(&tableName, &create); err != nil {
		return "", err
	}
	return create, nil
}

// ReverseAlterWithCompares 参照给定的对照, 生成给定的 AlterTableStmt 的反转语句
func ReverseAlterWithCompares(compares *ast.CreateTableStmt, alter *ast.AlterTableStmt) (string, bool, error) {
	if compares == nil {
		return "", false, errors.New("the raw CreateTableStmt is nill")
	}
	if alter == nil {
		return "", false, errors.New("the AlterTableStmt is nil")
	}

	if compares.Table.Name.String() != alter.Table.Name.String() {
		return "", false, errors.New("the altered table name is not equal with the compared table")
	}

	var restorer = ast.AlterTableStmt{
		Table: compares.Table,
		Specs: nil,
	}

	for _, spec := range alter.Specs {
		switch spec.Tp {
		// spec syntax:
		// tableOption [[,] tableOption]
		// processing caliber: take out the tableOption in the snapshot,
		// as long as the tableOption is modified,
		// roll it back to the snapshot state
		//
		// note: if the tableOption of the table is changed multiple times in a script,
		// the same rollback statement is generated multiple times and executed repeatedly,
		// but considering that repeated execution of the statement will not cause damage,
		// no better generation logic is sought.
		case ast.AlterTableOption:
			reverseSpec := *spec
			reverseSpec.Options = compares.Options
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		// processing caliber: drop the new column
		case ast.AlterTableAddColumns:
			for _, col := range spec.NewColumns {
				var reverseSpec ast.AlterTableSpec
				reverseSpec.Tp = ast.AlterTableDropColumn
				reverseSpec.OldColumnName = col.Name
				restorer.Specs = append(restorer.Specs, &reverseSpec)
			}

		// processing caliber: drop the new constraint
		case ast.AlterTableAddConstraint:
			var reverseSpec ast.AlterTableSpec
			reverseSpec.Tp = ast.AlterTableDropIndex
			reverseSpec.Name = spec.Constraint.Name
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		// not allowed operation: drop column
		case ast.AlterTableDropColumn:
			return "", false, errors.New("not allowed to AlterTableDropColumn")

		// not allowed operation: drop primary key
		case ast.AlterTableDropPrimaryKey:
			return "", false, errors.New("not allowed to AlterTableDropPrimaryKey")

		// processing caliber: add the index by AlterTableAddConstraint
		case ast.AlterTableDropIndex:
			if strings.EqualFold(spec.Name, "PRIMARY") {
				return "", false, errors.New("not allowed to AlterTableDropIndex PRIMARY")
			}

			var reverseSpec ast.AlterTableSpec
			reverseSpec.IfNotExists = true
			reverseSpec.Tp = ast.AlterTableAddConstraint

			for _, constraint := range compares.Constraints {
				if constraint.Name == spec.Name {
					reverseSpec.Constraint = constraint
					break
				}
			}

			if reverseSpec.Constraint != nil {
				restorer.Specs = append(restorer.Specs, &reverseSpec)
			}

		// processing caliber: forbid
		case ast.AlterTableDropForeignKey:
			return "", false, errors.New("not allowed to AlterTableDropForeignKey")

		// spec syntax:
		// MODIFY [ColumnKeywordOpt] [IfExists] {ColumnDef} [ColumnPosition]
		//
		// CHANGE [ColumnKeywordOpt] [IfExists] {ColumnName} {ColumnDef} [ColumnPosition]
		//
		// note: the renaming of columns by AlterTableChangeColumn is not considered here,
		// because column renaming is not compliant and will be filtered out in ErdaMySQLLint.
		// However, this processing loses versatility and will be repaired later.
		//
		// note: the change of AlterTableChangeColumn to ColumnPosition is not considered here.
		// the first reason is that the change to ColumnPosition will not affect the table structure,
		// the second reason is that the implementation of restoring the column position is slightly more complicated.
		//
		// ALTER [ColumnKeywordOpt] {ColumnName} SET DEFAULT  {SignedLiteral}
		//                                      │            └--(Expression)--┘
		//                                      └-DROP DEFAULT----------------┘
		//
		// processing caliber: whether it is SET DEFAULT or DROP DEFAULT, use AlterTableChangeColumn to fall back.
		case ast.AlterTableModifyColumn, ast.AlterTableChangeColumn, ast.AlterTableAlterColumn:
			var reverseSpec ast.AlterTableSpec
			reverseSpec.Tp = ast.AlterTableChangeColumn
			reverseSpec.IfExists = true
			reverseSpec.OldColumnName = spec.NewColumns[0].Name // 此处不考虑列被重命名的情况
			for _, col := range compares.Cols {
				if reverseSpec.OldColumnName.String() == col.Name.String() {
					reverseSpec.NewColumns = append(reverseSpec.NewColumns, col)
					break
				}
			}
			reverseSpec.Position = &ast.ColumnPosition{Tp: 0, RelativeColumn: nil}
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		// spec syntax:
		// RENAME COLUMN {identifier} TO {identifier}
		//
		// note: AlterTableRenameColumn will not appear under normal circumstances
		// it will be filtered out in ErdaMySQLLint.
		case ast.AlterTableRenameColumn:
			var reverseSpec ast.AlterTableSpec
			reverseSpec.Tp = ast.AlterTableRenameColumn
			reverseSpec.IfExists = true
			reverseSpec.NewColumnName = spec.OldColumnName
			reverseSpec.OldColumnName = spec.NewColumnName
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		// spec syntax:
		// RENAME------TO------{TableName}
		//      ├-----EqOpt----┤
		//		└------AS------┘
		//
		// note: AlterTableRenameTable will not appear under normal circumstances
		// it will be filtered out in ErdaMySQLLint.
		case ast.AlterTableRenameTable:
			var reverseSpec ast.AlterTableSpec
			reverseSpec.Tp = ast.AlterTableRenameTable
			reverseSpec.NewTable = restorer.Table
			restorer.Table = spec.NewTable
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		// processing caliber: forbid
		case ast.AlterTableLock:
			return "", false, errors.New("not allowed to AlterTableLock")

		// processing caliber: forbid
		case ast.AlterTableAlgorithm:
			return "", false, errors.New("not allowed to AlterTableAlgorithm")

		// spec syntax:
		// RENAME {INDEX|KEY} {identifier} TO {identifier}
		case ast.AlterTableRenameIndex:
			var reverseSpec ast.AlterTableSpec
			reverseSpec.Tp = ast.AlterTableRenameIndex
			reverseSpec.ToKey = spec.FromKey
			reverseSpec.FromKey = spec.ToKey
			restorer.Specs = append(restorer.Specs, &reverseSpec)

		case ast.AlterTableForce:
			return "", false, errors.New("not allowed to AlterTableForce")
		case ast.AlterTableAddPartitions:
			return "", false, errors.New("not allowed to AlterTableAddPartitions")
		case ast.AlterTableCoalescePartitions:
			return "", false, errors.New("not allowed to AlterTableCoalescePartitions")
		case ast.AlterTableDropPartition:
			return "", false, errors.New("not allowed to AlterTableDropPartition")
		case ast.AlterTableTruncatePartition:
			return "", false, errors.New("not allowed to AlterTableTruncatePartition")
		case ast.AlterTablePartition:
			return "", false, errors.New("not allowed to AlterTablePartition")
		case ast.AlterTableEnableKeys:
			return "", false, errors.New("not allowed to AlterTableEnableKeys")
		case ast.AlterTableDisableKeys:
			return "", false, errors.New("not allowed to AlterTableDisableKeys")
		case ast.AlterTableRemovePartitioning:
			return "", false, errors.New("not allowed to AlterTableRemovePartitioning")
		case ast.AlterTableWithValidation:
			return "", false, errors.New("not allowed to AlterTableWithValidation")
		case ast.AlterTableWithoutValidation:
			return "", false, errors.New("not allowed to AlterTableWithoutValidation")
		case ast.AlterTableSecondaryLoad:
			return "", false, errors.New("not allowed to AlterTableSecondaryLoad")
		case ast.AlterTableSecondaryUnload:
			return "", false, errors.New("not allowed to AlterTableSecondaryUnload")
		case ast.AlterTableRebuildPartition:
			return "", false, errors.New("not allowed to AlterTableRebuildPartition")
		case ast.AlterTableReorganizePartition:
			return "", false, errors.New("not allowed to AlterTableReorganizePartition")
		case ast.AlterTableCheckPartitions:
			return "", false, errors.New("not allowed to AlterTableCheckPartitions")
		case ast.AlterTableExchangePartition:
			return "", false, errors.New("not allowed to AlterTableExchangePartition")
		case ast.AlterTableOptimizePartition:
			return "", false, errors.New("not allowed to AlterTableOptimizePartition")
		case ast.AlterTableRepairPartition:
			return "", false, errors.New("not allowed to AlterTableRepairPartition")
		case ast.AlterTableImportPartitionTablespace:
			return "", false, errors.New("not allowed to AlterTableImportPartitionTablespace")
		case ast.AlterTableDiscardPartitionTablespace:
			return "", false, errors.New("not allowed to AlterTableDiscardPartitionTablespace")
		case ast.AlterTableAlterCheck:
			return "", false, errors.New("not allowed to AlterTableAlterCheck")
		case ast.AlterTableDropCheck:
			return "", false, errors.New("not allowed to AlterTableDropCheck")
		case ast.AlterTableImportTablespace:
			return "", false, errors.New("not allowed to AlterTableImportTablespace")
		case ast.AlterTableDiscardTablespace:
			return "", false, errors.New("not allowed to AlterTableDiscardTablespace")
		case ast.AlterTableIndexInvisible:
			return "", false, errors.New("not allowed to AlterTableIndexInvisible")
		case ast.AlterTableOrderByColumns:
			return "", false, errors.New("not allowed to AlterTableOrderByColumns")
		case ast.AlterTableSetTiFlashReplica:
			return "", false, errors.New("not allowed to AlterTableSetTiFlashReplica")
		}
	}

	if len(restorer.Specs) == 0 {
		return "", false, nil
	}
	var buf = bytes.NewBuffer(nil)
	if err := restorer.Restore(&format.RestoreCtx{
		Flags:     format.DefaultRestoreFlags,
		In:        buf,
		JoinLevel: 0,
	}); err != nil {
		return "", false, err
	}

	return buf.String(), buf.Len() != 0, nil
}

func ReverseDropIndexStmtWithCompares(compares *ast.CreateTableStmt, drop *ast.DropIndexStmt) (string, bool, error) {
	if strings.EqualFold(drop.IndexName, "PRIMARY") {
		return "", false, errors.New("not allowed to drop primary key")
	}

	var restorer = ast.AlterTableStmt{
		Table: compares.Table,
		Specs: make([]*ast.AlterTableSpec, 1),
	}

	var spec ast.AlterTableSpec
	spec.IfNotExists = true
	spec.Tp = ast.AlterTableAddConstraint

	for _, constraint := range compares.Constraints {
		if constraint.Name == drop.IndexName {
			spec.Constraint = constraint
			break
		}
	}
	if spec.Constraint == nil {
		return "", false, errors.Errorf("the index you droped not found, table name: %s, index name: %s", drop.Table.Name, drop.IndexName)
	}

	restorer.Specs[0] = &spec
	var buf = bytes.NewBuffer(nil)
	if err := restorer.Restore(&format.RestoreCtx{
		Flags:     format.DefaultRestoreFlags,
		In:        buf,
		JoinLevel: 0,
	}); err != nil {
		return "", false, errors.Wrap(err, "failed to Restore SQL node")
	}

	return buf.String(), buf.String() != "", nil
}

func ReverseCreateTableStmt(in *ast.CreateTableStmt) string {
	if in == nil || in.Table == nil {
		return ""
	}
	tableName := in.Table.Name.String()
	return "DROP TABLE IF EXISTS " + tableName + ";\n"
}
