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
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (mig *Migrator) DB() *gorm.DB {
	if mig.db != nil {
		return mig.db
	}

	var (
		err error
		dsn = mig.MySQLParameters().Format(false)
	)
	open, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Fatalf("failed to open MySQL connection with DSN %s, err: %v", dsn, err)
	}
	defer open.Close()

	stmt, err := open.Prepare("CREATE DATABASE IF NOT EXISTS " + mig.dbSettings.Name + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
	if err != nil {
		logrus.Fatalf("failed to Prepare CREATE DATABASE stmt, err: %v", err)
	}
	_, err = stmt.Exec()
	if err != nil {
		logrus.Fatalf("failed to Exec stmt %+v, err: %v", stmt, err)
	}

	dsn = mig.MySQLParameters().Format(true)
	mig.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.Fatalf("failed to open MySQL connection with DSN %s : %v", dsn, err)
	}
	mig.db.Logger = logger.New(
		log.New(os.Stdout, "\r\n", log.Ltime),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Error,
		},
	)

	if mig.Parameters.DebugSQL() {
		mig.db = mig.db.Debug()
	}

	return mig.db
}

// ClearSandbox
// if you want to do a new migration in a clean sandbox,
// you should ClearSandbox to clear all changes on last migration
func (mig *Migrator) ClearSandbox() {
	mig.sandbox = nil
}

func (mig *Migrator) SandBox() *gorm.DB {
	if mig.sandbox != nil {
		return mig.sandbox
	}

	var (
		open           *sql.DB
		create         *sql.Stmt
		err            error
		createDatabase = "CREATE DATABASE IF NOT EXISTS " + mig.sandboxSettings.Name + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
		dropDatabase   = "DROP SCHEMA IF EXISTS " + mig.sandboxSettings.Name
	)
	defer func() {
		if open != nil {
			_ = open.Close()
		}
	}()

	var (
		timeout = time.Second * 150
		dsn     = mig.SandboxParameters().Format(false)
	)

	for now := time.Now(); time.Since(now) < timeout; time.Sleep(time.Second * 3) {
		open, err = sql.Open("mysql", dsn)
		if err != nil {
			logrus.WithField("DSN", dsn).WithError(err).Infoln("failed to connect to sandbox")
			continue
		}

		drop, err := open.Prepare(dropDatabase)
		if err != nil {
			logrus.Warnf("failed to Prepare %s stmt: %v, may sandbox is not working yet, wait it for %d seconds", dropDatabase, err, timeout.Seconds()-time.Since(now).Seconds())
			continue
		}
		if _, err := drop.Exec(); err != nil {
			logrus.Warnf("failed to Exec %s: %v", dropDatabase, err)
		}

		create, err = open.Prepare(createDatabase)
		if err != nil {
			logrus.Warnf("failed to Prepare %s stmt: %v, may sandbox is not working yet, wait it for %v sencods", createDatabase, err, int(timeout.Seconds()-time.Since(now).Seconds()))
			continue
		}

		if _, err = create.Exec(); err != nil {
			logrus.Fatalf("failed to Exec prepared %s stmt, err: %v", strconv.Quote(createDatabase), err)
		}
		break
	}
	if err != nil {
		logrus.Fatalf("failed to dial MySQL sandbox: %s: %v", dsn, err)
	}

	dsn = mig.SandboxParameters().Format(true)
	mig.sandbox, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.Fatalf("failed to open connection to sandbox, sandbox: %s: %v", mig.SandboxParameters().Format(true), err)
	}
	mig.sandbox.Logger = logger.New(
		log.New(os.Stdout, "\r\n", log.Ltime),
		logger.Config{SlowThreshold: 200 * time.Millisecond, LogLevel: logger.Error, Colorful: true},
	)

	if mig.Parameters.DebugSQL() {
		mig.sandbox = mig.sandbox.Debug()
	}

	return mig.sandbox
}
