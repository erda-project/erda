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
		err  error
		dsn  = mig.MySQLParameters().Format(false)
		stmt = "CREATE DATABASE IF NOT EXISTS " + mig.dbSettings.Name + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
	)
	open, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.WithError(err).WithField("DSN", dsn).Fatalln("failed to open MySQL connection")
	}
	defer open.Close()

	if _, err = open.Exec(stmt); err != nil {
		logrus.WithError(err).Fatalf("failed to Exec stmt %+v", stmt)
	}

	dsn = mig.MySQLParameters().Format(true)
	mig.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.WithError(err).WithField("DSN", dsn).Fatalln("failed to open MySQL connection")
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
		waiting := timeout.Seconds() - time.Since(now).Seconds()
		open, err = sql.Open("mysql", dsn)
		if err != nil {
			logrus.WithError(err).WithField("DSN", dsn).
				Warnf("failed to connect to sandbox, may it is not working yet, wait it for %.1f seconds", waiting)
			continue
		}

		if _, err = open.Exec(dropDatabase); err != nil {
			logrus.WithError(err).WithField("SQL", dropDatabase).
				Warnf("failed to Exec, may the sandbox it not working yet, wait it for %.1f seconds", waiting)
			continue
		}

		if _, err = open.Exec(createDatabase); err != nil {
			logrus.WithError(err).WithField("SQL", createDatabase).Fatalln("failed to Exec")
			continue
		}

		break
	}
	if err != nil {
		logrus.WithError(err).WithField("DSN", dsn).Fatalln("failed to dial MySQL sandbox")
	}

	dsn = mig.SandboxParameters().Format(true)
	mig.sandbox, err = gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.WithError(err).WithField("DSN", dsn).Fatalln("failed to open connection to sandbox")
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
