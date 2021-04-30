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

package migrator

import (
	"database/sql"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dsn        string
	sandboxDSN string
	database   string
	db         *gorm.DB
	sandbox    *gorm.DB
	debugSQL   bool
)

func DB() *gorm.DB {
	if db == nil {
		if dsn == "" {
			logrus.Fatalln("MySQL DSN is invalid, did you input the right DSN in command line args ?")
		}

		var err error
		open, err := sql.Open("mysql", dsn)
		if err != nil {
			logrus.Fatalf("failed to open MySQL connection with DSN %s, err: %v", dsn, err)
		}
		defer open.Close()

		stmt, err := open.Prepare("CREATE DATABASE IF NOT EXISTS " + database + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			logrus.Fatalf("failed to Prepare CREATE DATABASE stmt, err: %v", err)
		}
		_, err = stmt.Exec()
		if err != nil {
			logrus.Fatalf("failed to Exec stmt %+v, err: %v", stmt, err)
		}

		values := make(url.Values, 2)
		values.Add("timeout", "150s")
		values.Add("parseTime", "True")
		dsn += database + "?" + values.Encode()
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
		if err != nil {
			logrus.Fatalf("failed to open MySQL connection with DSN %s : %v", strconv.Quote(dsn), err)
		}
		db.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.Ltime),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				Colorful:                  true,
				IgnoreRecordNotFoundError: true,
				LogLevel:                  logger.Error,
			},
		)

		if debugSQL {
			db = db.Debug()
		}
	}

	return db
}

// ClearSandbox
// if you want to do a new migration in a clean sandbox,
// you should ClearSandbox to clear all changes on last migration
func ClearSandbox() {
	sandbox = nil
}

func SandBox() *gorm.DB {
	if sandboxDSN == "" {
		logrus.Fatalln("sandbox DSN is invalid")
	}

	if sandbox == nil {
		var (
			open           *sql.DB
			create         *sql.Stmt
			err            error
			createDatabase = "CREATE DATABASE IF NOT EXISTS " + database + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
			dropDatabase   = "DROP SCHEMA IF EXISTS " + database
		)
		defer func() {
			if open != nil {
				_ = open.Close()
			}
		}()
		for now := time.Now(); time.Since(now) < time.Second*150; time.Sleep(time.Second * 3) {
			open, err = sql.Open("mysql", sandboxDSN)
			if err != nil {
				logrus.Infof("failed to connect to MySQL sandbox, DSN: %s, err: %v", sandboxDSN, err)
				continue
			}

			drop, err := open.Prepare(dropDatabase)
			if err != nil {
				logrus.Warnf("failed to Prepare %s stmt: %v", dropDatabase, err)
				continue
			}
			if _, err := drop.Exec(); err != nil {
				logrus.Warnf("failed to Exec %s: %v", dropDatabase, err)
			}

			create, err = open.Prepare(createDatabase)
			if err != nil {
				logrus.Warnf("failed to Prepare %s stmt, err: %v", strconv.Quote(createDatabase), err)
				continue
			}

			if _, err = create.Exec(); err != nil {
				logrus.Fatalf("failed to Exec prepared %s stmt, err: %v", strconv.Quote(createDatabase), err)
			}
			break
		}
		if err != nil {
			logrus.Fatalf("failed to dial MySQL sandbox, DSN: %s, err: %v", sandboxDSN, err)
		}

		values := make(url.Values, 2)
		values.Add("timeout", "150s")
		values.Add("parseTime", "True")
		sandboxDSN += database + "?" + values.Encode()
		sandbox, err = gorm.Open(mysql.Open(sandboxDSN), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
		if err != nil {
			logrus.Fatalf("failed to open connection to sandbox, sandbox dsn: %s, err: %v", sandboxDSN, err)
		}
		sandbox.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.Ltime),
			logger.Config{SlowThreshold: 200 * time.Millisecond, LogLevel: logger.Error, Colorful: true},
		)

		if debugSQL {
			sandbox = sandbox.Debug()
		}
	}

	return sandbox
}

// SchemaMigrationHistoryExists returns whether schema_migration_history exists
func SchemaMigrationHistoryExists() bool {
	return Raw("SHOW TABLES LIKE '?'", SchemaMigrationHistory).RowsAffected == 1
}

// ShowCreateTables select table's DDL
// if tableNames is nil returns all tables DDL
func ShowCreateTables(tableNames ...string) (creates []string, err error) {
	if len(tableNames) == 0 {
		return showCreateTables()
	}

	for _, tableName := range tableNames {
		var create string
		tx := Raw("show create table " + tableName)
		if err = tx.Error; err != nil {
			return nil, err
		}
		if err = tx.Row().Scan(&tableName, &create); err != nil {
			return nil, err
		}
		creates = append(creates, create)
	}
	return creates, nil
}

func showCreateTables() ([]string, error) {
	tx := Raw("SHOW TABLES")
	if err := tx.Error; err != nil {
		return nil, err
	}
	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}

	var creates []string
	for rows.Next() {
		var (
			tableName string
			create    string
		)
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tx := Raw("SHOW CREATE TABLE " + tableName)
		if err := tx.Error; err != nil {
			return nil, err
		}
		if err := tx.Row().Scan(&tableName, &create); err != nil {
			return nil, err
		}
		creates = append(creates, create)
	}

	return creates, nil
}

func Begin() *gorm.DB                                           { return DB().Begin() }
func First(dest interface{}, cond ...interface{}) (tx *gorm.DB) { return DB().First(dest, cond...) }
func Find(dest interface{}, cond ...interface{}) (tx *gorm.DB)  { return DB().Find(dest, cond...) }
func Raw(sql string, values ...interface{}) (tx *gorm.DB)       { return DB().Raw(sql, values...) }
func Exec(sql string, values ...interface{}) (tx *gorm.DB)      { return DB().Exec(sql, values...) }
