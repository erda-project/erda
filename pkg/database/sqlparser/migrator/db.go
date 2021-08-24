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
		err         error
		timeout     = time.Second * 360
		dsn         = mig.MySQLParameters().Format(false)
		showSchemas = "SHOW SCHEMAS LIKE ?"
		stmt        = "CREATE DATABASE IF NOT EXISTS " + mig.dbSettings.Name + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
	)

	// initF shows schemas like the database, if the database is not exists, create it
	initF := func(db *sql.DB) error {
		row := db.QueryRow(showSchemas, mig.MySQLParameters().Database)
		if row != nil {
			return nil
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
		return nil
	}
	if err = mig.initConnToDB(dsn, timeout, initF); err != nil {
		logrus.WithError(err).Fatalln("failed to init connection to MySQL Server")
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
			LogLevel:                  logger.Silent,
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
		err            error
		createDatabase = "CREATE DATABASE IF NOT EXISTS " + mig.sandboxSettings.Name + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
		dropDatabase   = "DROP SCHEMA IF EXISTS " + mig.sandboxSettings.Name
		timeout        = time.Second * 150
		dsn            = mig.SandboxParameters().Format(false)
	)

	initF := func(db *sql.DB) error {
		if _, err := db.Exec(dropDatabase); err != nil {
			return err
		}
		if _, err := db.Exec(createDatabase); err != nil {
			return err
		}
		return nil
	}
	if err = mig.initConnToDB(dsn, timeout, initF); err != nil {
		logrus.WithError(err).Fatalln("failed to init connection to the sandbox")
		return nil
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

func (mig *Migrator) initConnToDB(dsn string, timeout time.Duration, initF func(db *sql.DB) error) (err error) {
	var (
		open *sql.DB
	)
	defer func() {
		if open != nil {
			_ = open.Close()
		}
	}()

	for now := time.Now(); time.Since(now) < timeout; time.Sleep(time.Second * 3) {
		open, err = sql.Open("mysql", dsn)
		if err != nil {
			logrus.WithError(err).WithField("DSN", dsn).WithField("left time", timeout.Seconds()-time.Since(now).Seconds()).
				Warnln("failed to connect to the MySQL Server, it will try again in 3 seconds")
			continue
		}
		if err := open.Ping(); err != nil {
			continue
		}
		if err = initF(open); err != nil {
			return err
		}

		return nil
	}

	logrus.WithError(err).WithField("DSN", dsn).Fatalln("failed to dial the MySQL Server")
	return err
}
