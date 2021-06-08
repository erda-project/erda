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
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/sqlparser/ddlreverser"
	"github.com/erda-project/erda/pkg/database/sqlparser/snapshot"
)

const defaultDatabase = "dice"

// installing type
const (
	firstTimeInstall = "first_time_install"
	normalUpdate     = "normal_update"
	firstTimeUpdate  = "first_time_update"
)

type Migrator struct {
	Parameters

	snap         *snapshot.Snapshot
	LocalScripts *Scripts
	reversing    []string

	installingType string
}

func New(parameters Parameters) (*Migrator, error) {
	if parameters == nil {
		return nil, errors.New("parameters is nil")
	}

	// init parameters
	dsn = parameters.DSN()
	sandboxDSN = parameters.SandboxDSN()
	debugSQL = parameters.DebugSQL()
	database = parameters.Database()
	if database == "" {
		database = defaultDatabase
	}

	scripts, err := NewScripts(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewScripts")
	}

	var mig = Migrator{
		Parameters:   parameters,
		snap:         nil,
		LocalScripts: scripts,
		reversing:    nil,
	}
	return &mig, nil
}

func (mig *Migrator) Run() error {
	// snapshot database schema structure
	snap, err := snapshot.From(DB(), SchemaMigrationHistory)
	if err != nil {
		return err
	}
	mig.snap = snap

	// if there is no table then goto new installation
	if !snap.HasAnyTable() {
		logrus.Infoln("first time installation...")
		mig.installingType = firstTimeInstall
		return mig.newInstallation()
	}
	// if there is any history record then goto normal update,
	// otherwise goto first time update
	tx := DB().Find(new([]HistoryModel))
	if tx.Error == nil && tx.RowsAffected > 0 {
		logrus.Infoln("normal update...")
		mig.installingType = normalUpdate
		return mig.normalUpdate()
	}
	if tx.Error != nil {
		logrus.Infof("failed to Find HistoryModel records, err: %v", tx.Error)
	}

	logrus.Infoln("first time update...")
	mig.installingType = firstTimeUpdate
	return mig.firstTimeUpdate()
}

func (mig *Migrator) newInstallation() (err error) {
	history := new(HistoryModel)
	history.create()

	// Erda mysql lint
	if mig.NeedErdaMySQLLint() {
		logrus.Infoln("DO ERDA MYSQL LINT...")
		if err = mig.LocalScripts.Lint(); err != nil {
			return err
		}
		logrus.Infoln("ERDA MYSQL LINT OK")
	}

	// alter permission lint
	logrus.Infoln("DO ALTER PERMISSION LINT...")
	if err = mig.LocalScripts.AlterPermissionLint(); err != nil {
		return err
	}
	logrus.Infoln("ALTER PERMISSION LINT OK")
	// execute in sandbox
	logrus.Infoln("DO MIGRATION IN SANDBOX...")
	if err = mig.migrateSandbox(); err != nil {
		return err
	}
	logrus.Infoln("MIGRATE IN SANDBOX OK")

	// migrate schema SQLs
	logrus.Infoln("DO PRE-MIGRATION...")
	if err = mig.preMigrate(); err != nil {
		return err
	}
	logrus.Infoln("PRE-MIGRATE OK")

	// migrate data SQLs
	logrus.Infoln("DO MIGRATION...")
	if err = mig.migrate(); err != nil {
		return err
	}
	logrus.Infoln("MIGRATE OK")

	return nil
}

func (mig *Migrator) normalUpdate() (err error) {
	// Erda mysql lint
	if mig.NeedErdaMySQLLint() {
		logrus.Infoln("DO ERDA MYSQL LINT....")
		if err = mig.LocalScripts.Lint(); err != nil {
			return err
		}
		logrus.Infoln("ERDA MYSQL LINT OK")
	}

	// same name lint
	logrus.Infoln("DO SAME NAME LINT....")
	if err = mig.LocalScripts.SameNameLint(); err != nil {
		return err
	}
	logrus.Infoln("SAME NAME LINT OK")

	// alter permission lint
	logrus.Infoln("DO ALTER PERMISSION LINT....")
	if err = mig.LocalScripts.AlterPermissionLint(); err != nil {
		return err
	}
	logrus.Infoln("ALTER PERMISSION LINT OK")

	// installed script changes lint
	logrus.Infoln("DO INSTALLED CHANGES LINT....")
	if err = mig.LocalScripts.InstalledChangesLint(); err != nil {
		return err
	}
	logrus.Infoln("INSTALLED CHANGES LINT OK")

	// copy database snapshot to sandbox
	logrus.Infoln("COPY CURRENT DATABASE STRUCTURE TO SANDBOX....")
	if err = mig.snap.RecoverTo(SandBox()); err != nil {
		return err
	}
	logrus.Infoln("COPY CURRENT DATABASE STRUCTURE TO SANDBOX OK")

	// migrate in sandbox
	logrus.Infoln("DO MIGRATION IN SANDBOX....")
	if err = mig.migrateSandbox(); err != nil {
		return err
	}
	logrus.Infoln("MIGRATE IN SANDBOX OK")

	// pre migrate data
	logrus.Infoln("DO PRE-MIGRATION....")
	if err = mig.preMigrate(); err != nil {
		return err
	}
	logrus.Infoln("PRE-MIGRATE OK")

	// migrate data
	logrus.Infoln("DO MIGRATION....")
	if err = mig.migrate(); err != nil {
		return err
	}
	logrus.Infoln("MIGRATE OK")

	return nil
}

func (mig *Migrator) firstTimeUpdate() (err error) {
	// the correct state in this time is db.Schema == baseScriptsSchema,
	// if not, returns error, manual intervention.
	// marks all baseScripts !pending, marks others pending

	// compare local base schema and cloud base schema for every service
	logrus.Infoln("COMPARE LOCAL SCHEMA AND CLOUD SCHEMA FOR EVERY SERVICE... ..")
	for _, service := range mig.LocalScripts.Services {
		if equal := service.BaselineEqualCloud(DB()); !equal.Equal() {
			return errors.Errorf("local base schema is not equal with cloud schema: %s", equal.Reason())
		}
	}

	// record base
	logrus.Infoln("RECORD BASE... ..")
	now := time.Now()
	new(HistoryModel).create()
	for serviceName, service := range mig.LocalScripts.Services {
		for i := range service {
			if service[i].isBase {
				service[i].Pending = false
				if err := (&HistoryModel{
					ID:           0,
					CreatedAt:    now,
					UpdatedAt:    now,
					ServiceName:  serviceName,
					Filename:     filepath.Base(service[i].Name),
					Checksum:     service[i].Checksum(),
					InstalledBy:  "",
					InstalledOn:  "",
					LanguageType: "SQL",
					Reversed:     ddlreverser.ReverseCreateTableStmtsToDropTableStmts(service[i]),
				}).insert(); err != nil {
					return err
				}
			}
		}
	}

	return mig.normalUpdate()
}

func (mig *Migrator) reverse(reversing []string, reverseSlice bool) {
	if len(reversing) == 0 {
		return
	}
	if reverseSlice {
		ddlreverser.ReverseSlice(reversing)
	}
	for _, s := range mig.reversing {
		if err := DB().Exec(s).Error; err != nil {
			logrus.Fatalf("failed to exec reversing SQL: %s", s)
		}
	}
}

func (mig *Migrator) migrateSandbox() (err error) {
	for _, serviceName := range mig.LocalScripts.ServicesNames {
		scripts := mig.LocalScripts.Services[serviceName]
		for _, script := range scripts {
			if !script.Pending {
				continue
			}
			for _, ddl := range script.DDLNodes() {
				if err = SandBox().Exec(ddl.Text()).Error; err != nil {
					return errors.Wrapf(err, "failed to migrate in sandbox. The service name: %s, the script filename: %s, the schema SQL: %s",
						serviceName, script.Name, ddl.Text())
				}
			}
			for _, dml := range script.DMLNodes() {
				if err = SandBox().Exec(dml.Text()).Error; err != nil {
					return errors.Wrapf(err, "failed to migrate in sandbox. The service name: %s, the script filename: %s, the data SQL: %s",
						serviceName, script.Name, dml.Text())
				}
			}
		}
	}
	return err
}

// pre migrate data SQLs, all applied in this runtime will be rollback
func (mig *Migrator) preMigrate() error {
	if mig.LocalScripts.HasDestructiveOperationInPending() {
		logrus.Warnln("there are destructive SQL in pending scripts, stop doing pre-migration")
		return nil
	}

	// finally roll all DDL back
	defer func() {
		logrus.Infoln("\tREVERSE PRE-MIGRATIONS")
		mig.reverse(mig.reversing, true)
	}()

	mig.reversing = nil

	for _, serviceName := range mig.LocalScripts.ServicesNames {
		scripts := mig.LocalScripts.Services[serviceName]
		for _, script := range scripts {
			if !script.Pending {
				continue
			}
			if err := script.Install(Begin, func(tx *gorm.DB, _ error) {
				logrus.Infof("\tROLLBACK PRE-MIGRATIONS: %s", script.Name)
				tx.Rollback()
			}); err != nil {
				return err
			}

			mig.reversing = append(mig.reversing, script.Reversing...)
		}
	}

	return nil
}

func (mig *Migrator) migrate() error {
	now := time.Now()
	for _, serviceName := range mig.LocalScripts.ServicesNames {
		scripts := mig.LocalScripts.Services[serviceName]
		for _, script := range scripts {
			if !script.Pending {
				continue
			}
			logrus.Infoln("install", script.Name)
			if err := script.Install(Begin, func(tx *gorm.DB, err error) {
				if err != nil {
					tx.Rollback()
					mig.reverse(script.Reversing, true) // script granularity rollback, not overall rollback.
				} else {
					tx.Commit()
				}
			}); err != nil {
				return err
			}

			ddlreverser.ReverseSlice(script.Reversing)
			if err := (&HistoryModel{
				ID:           0,
				CreatedAt:    now,
				UpdatedAt:    now,
				ServiceName:  serviceName,
				Filename:     filepath.Base(script.Name),
				Checksum:     script.Checksum(),
				InstalledBy:  "",
				InstalledOn:  "",
				LanguageType: "SQL",
				Reversed:     strings.Join(script.Reversing, "\n"),
			}).insert(); err != nil {
				return errors.Wrap(err, "internal error: failed to record migration. All migrations will be rolled back.")
			}
		}
	}

	return nil
}
