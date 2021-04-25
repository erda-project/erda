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

package ddlreverser_test

import (
	"bytes"
	"flag"
	"testing"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	"github.com/pingcap/parser/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/sqlparser/ddlreverser"

	_ "github.com/pingcap/tidb/types/parser_driver"
)

const createTable = `
CREATE TABLE IF NOT EXISTS dice_api_doc_lock (
  id bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  session_id char(36) NOT NULL COMMENT '会话标识',
  is_locked tinyint(1) NOT NULL DEFAULT '0' COMMENT '会话所有者是否持有文档锁',
  expired_at datetime NOT NULL COMMENT '会话过期时间',
  application_id bigint(20) NOT NULL COMMENT '应用 id',
  branch_name varchar(191) NOT NULL COMMENT '分支名',
  doc_name varchar(191) NOT NULL COMMENT '文档名, 也即服务名',
  creator_id varchar(191) NOT NULL COMMENT '创建者 id',
  updater_id varchar(191) NOT NULL COMMENT '更新者 id',
  PRIMARY KEY (id),
  UNIQUE KEY uk_doc (application_id, branch_name,doc_name)
) ENGINE=InnoDB AUTO_INCREMENT=399 DEFAULT CHARSET=utf8mb4 COMMENT='API 设计中心文档锁表';

RENAME TABLE dice_api_doc_lock to doc_lock;

ALTER TABLE doc_lock
	COMMENT='修改了的 table option';

ALTER TABLE doc_lock
	ADD (
		new_col1 varchar(16) not null comment 'this is a new column',
		new_col2 bigint not null
	);

ALTER TABLE doc_lock 
	ADD CONSTRAINT new_constraint UNIQUE KEY (doc_name);

ALTER TABLE doc_lock
	MODIFY new_col1 varchar(1024) default 'default text' comment 'modified the column';

ALTER TABLE doc_lock
	CHANGE new_col1 new_col1 text not null comment 'change the column';

ALTER TABLE doc_lock
	RENAME COLUMN new_col2 to new_col3;

ALTER TABLE doc_lock
	ALTER new_col3 set default 99;

ALTER TABLE doc_lock
	RENAME INDEX uk_doc TO uk_doc2;
`

var dsn = flag.String("dsn", "", "mysql dsn")

func openDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(*dsn))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestReverseDDLWithSnapshot(t *testing.T) {
	if dsn == nil || *dsn == "" {
		return
	}
	db, err := openDB()
	if err != nil {
		t.Fatalf("failed to openDB, err: %v", err)
	}

	defer db.Migrator().DropTable("dice_api_doc_lock", "doc_lock")

	nodes, warns, err := parser.New().Parse(createTable, "", "")
	if err != nil {
		t.Fatalf("failed to Pasrse, err: %v", err)
	}
	for _, warn := range warns {
		t.Log("warn:", warn)
	}

	for i, node := range nodes {
		t.Logf("node text: %s", node.Text())
		reversed, _, err := ddlreverser.ReverseDDLWithSnapshot(db, node.(ast.DDLNode))
		if err != nil {
			t.Fatalf("failed to ReverseDDLWithSnapshot [%v], err: %v", i, err)
		}
		t.Logf("reversed SQL: %s", reversed)

		if err := db.Exec(node.Text()).Error; err != nil {
			t.Fatalf("failed to Exec [%v]: nerr: %v", i, err)
		}
	}

}

const (
	createWithIndex = `
CREATE TABLE dice_api_doc_lock (
  id bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  session_id char(36) NOT NULL COMMENT '会话标识',
  is_locked tinyint(1) NOT NULL DEFAULT '0' COMMENT '会话所有者是否持有文档锁',
  expired_at datetime NOT NULL COMMENT '会话过期时间',
  application_id bigint(20) NOT NULL COMMENT '应用 id',
  branch_name varchar(191) NOT NULL COMMENT '分支名',
  doc_name varchar(191) NOT NULL COMMENT '文档名, 也即服务名',
  creator_id varchar(191) NOT NULL COMMENT '创建者 id',
  updater_id varchar(191) NOT NULL COMMENT '更新者 id',
  PRIMARY KEY (id),
  UNIQUE KEY uk_doc (application_id,branch_name,doc_name)
) ENGINE=InnoDB AUTO_INCREMENT=418 DEFAULT CHARSET=utf8mb4 COMMENT='API 设计中心文档锁表'
`
	alterTableDropIndex = "ALTER TABLE dice_api_doc_lock DROP INDEX uk_doc"
	dropIndex           = "DROP INDEX `PRIMARY` ON dice_api_doc_lock"
	dropIndex2          = "DROP INDEX uk_doc ON dice_api_doc_lock"
)

var constraintTypes = map[ast.ConstraintType]string{
	ast.ConstraintNoConstraint: "ConstraintNoConstraint",
	ast.ConstraintPrimaryKey:   "ConstraintPrimaryKey",
	ast.ConstraintKey:          "ConstraintKey",
	ast.ConstraintIndex:        "ConstraintIndex",
	ast.ConstraintUniq:         "ConstraintUniq",
	ast.ConstraintUniqKey:      "ConstraintUniqKey",
	ast.ConstraintUniqIndex:    "ConstraintUniqIndex",
	ast.ConstraintForeignKey:   "ConstraintForeignKey",
	ast.ConstraintFulltext:     "ConstraintFulltext",
	ast.ConstraintCheck:        "ConstraintCheck",
}

func TestReverseAlterWithCompares_AlterTableDropIndex(t *testing.T) {
	node, err := parser.New().ParseOneStmt(createWithIndex, "", "")
	if err != nil {
		t.Error(err)
	}
	createNode := node.(*ast.CreateTableStmt)

	for _, constraint := range createNode.Constraints {
		var buf = bytes.NewBuffer(nil)
		_ = constraint.Restore(&format.RestoreCtx{
			Flags:     format.DefaultRestoreFlags,
			In:        buf,
			JoinLevel: 0,
		})
		t.Logf("%+v, constraint.Tp: %s, restored: %s", *constraint, constraintTypes[constraint.Tp], buf.String())

		stmt := ast.AlterTableStmt{
			Table: createNode.Table,
			Specs: []*ast.AlterTableSpec{{
				IfExists:        true,
				IfNotExists:     true,
				NoWriteToBinlog: false,
				OnAllPartitions: false,
				Tp:              ast.AlterTableAddConstraint,
				Name:            "",
				Constraint:      constraint,
				Options:         nil,
				OrderByList:     nil,
				NewTable:        nil,
				NewColumns:      nil,
				NewConstraints:  nil,
				OldColumnName:   nil,
				NewColumnName:   nil,
				Position:        nil,
				LockType:        0,
				Algorithm:       0,
				Comment:         "",
				FromKey:         model.CIStr{},
				ToKey:           model.CIStr{},
				Partition:       nil,
				PartitionNames:  nil,
				PartDefinitions: nil,
				WithValidation:  false,
				Num:             0,
				Visibility:      0,
				TiFlashReplica:  nil,
			}},
		}
		var buf2 = bytes.NewBuffer(nil)
		if err := stmt.Restore(&format.RestoreCtx{
			Flags:     format.DefaultRestoreFlags,
			In:        buf2,
			JoinLevel: 0,
		}); err != nil {
			t.Fatal(err)
		}

		t.Logf("reversed: %s", buf2.String())
	}

	alterTableDropIndexStmtNode, err := parser.New().ParseOneStmt(alterTableDropIndex, "", "")
	if err != nil {
		t.Error(err)
	}
	t.Logf("alterTableDropIndexStmtNode: %+v, specs[0]: %+v", alterTableDropIndexStmtNode, alterTableDropIndexStmtNode.(*ast.AlterTableStmt).Specs[0])

	dropIndexStmtNode, err := parser.New().ParseOneStmt(dropIndex, "", "")
	if err != nil {
		t.Error(err)
	}
	t.Logf("dropIndexStmtNode: %+v", dropIndexStmtNode)

	dropIndexStmtNode2, err := parser.New().ParseOneStmt(dropIndex2, "", "")
	if err != nil {
		t.Error(err)
	}
	t.Logf("dropIndexStmtNode2: %+v", dropIndexStmtNode2)

}
