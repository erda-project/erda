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
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/sqlparser/ddlreverser"
	"github.com/erda-project/erda/pkg/database/sqlparser/pygrator"
	"github.com/erda-project/erda/pkg/database/sqlparser/snapshot"
)

// installing type
const (
	firstTimeInstall installType = "first_time_install"
	normalUpdate     installType = "normal_update"
	firstTimeUpdate  installType = "first_time_update"
)

type installType string

type Migrator struct {
	Parameters

	snap         *snapshot.Snapshot
	LocalScripts *Scripts
	reversing    []string

	installingType installType

	dbSettings      *pygrator.Settings // database settings
	sandboxSettings *pygrator.Settings // sandbox settings
	db              *gorm.DB
	sandbox         *gorm.DB
}

func New(parameters Parameters) (mig *Migrator, err error) {
	if parameters == nil {
		return nil, errors.New("parameters is nil")
	}
	if parameters.MySQLParameters() == nil {
		return nil, errors.New("MySQL DSN is invalid, did you set the right DSN ?")
	}
	if parameters.SandboxParameters() == nil {
		return nil, errors.New("sandbox DSN is invalid, did you set the right DSN ?")
	}

	// init parameters
	mig = new(Migrator)
	mig.Parameters = parameters
	mig.dbSettings = &pygrator.Settings{
		Engine:   pygrator.DjangoMySQLEngine,
		User:     mig.MySQLParameters().Username,
		Password: mig.MySQLParameters().Password,
		Host:     mig.MySQLParameters().Host,
		Port:     mig.MySQLParameters().Port,
		Name:     mig.MySQLParameters().Database,
		TimeZone: pygrator.TimeZoneAsiaShanghai,
	}
	mig.sandboxSettings = &pygrator.Settings{
		Engine:   pygrator.DjangoMySQLEngine,
		User:     mig.SandboxParameters().Username,
		Password: mig.SandboxParameters().Password,
		Host:     mig.SandboxParameters().Host,
		Port:     mig.SandboxParameters().Port,
		Name:     mig.SandboxParameters().Database,
		TimeZone: pygrator.TimeZoneAsiaShanghai,
	}

	// load scripts
	mig.LocalScripts, err = NewScripts(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewScripts")
	}

	return mig, nil
}

func (mig *Migrator) Run() (err error) {
	defer func() {
		if err == nil {
			logrus.Infoln("Erda MySQL Migrate Complete !")
		}
	}()

	// snapshot database schema structure
	mig.snap, err = snapshot.From(mig.DB(), SchemaMigrationHistory)
	if err != nil {
		return err
	}

	// if there is no table then goto new installation
	if !mig.snap.HasAnyTable() {
		logrus.Infoln("there is not any table, it is the first time installation...")
		mig.installingType = firstTimeInstall
		return mig.newInstallation()
	}

	// if there is any history record then goto normal update,
	// otherwise goto first time update
	tx := mig.DB().Find(new([]HistoryModel))
	if tx.Error == nil && tx.RowsAffected > 0 {
		logrus.Infoln("found migration histries, it is the normal update installation...")
		mig.installingType = normalUpdate
		return mig.normalUpdate()
	}

	if tx.Error != nil {
		logrus.WithError(tx.Error).Warnln("failed to Find HistoryModel records")
	}
	logrus.Infoln("there are some tables but no migration histories, it is the first time update installation by Erda MySQL Migration...")
	mig.installingType = firstTimeUpdate
	return mig.firstTimeUpdate()
}

func (mig *Migrator) newInstallation() (err error) {
	new(HistoryModel).CreateTable(mig.DB())
	mig.LocalScripts.MarkPending(mig.DB())

	// Erda mysql lint
	if !mig.SkipMigrationLint() {
		logrus.Infoln("DO ERDA MYSQL LINT...")
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
	logrus.Infoln("DO ALTER PERMISSION LINT...")
	if err = mig.LocalScripts.AlterPermissionLint(); err != nil {
		return err
	}
	logrus.Infoln("ALTER PERMISSION LINT OK")

	// execute in sandbox
	if !mig.SkipSandbox() {
		logrus.Infoln("DO MIGRATION IN SANDBOX...")
		if err = mig.migrateSandbox(); err != nil {
			return err
		}
		logrus.Infoln("MIGRATE IN SANDBOX OK")
	}

	// pre-migrate schema SQLs
	if !mig.SkipPreMigrate() && !mig.SkipMigrate() {
		logrus.Infoln("DO PRE-MIGRATION...")
		if err = mig.preMigrate(); err != nil {
			return err
		}
		logrus.Infoln("PRE-MIGRATE OK")
	}

	// migrate data SQLs
	if !mig.SkipMigrate() {
		logrus.Infoln("DO MIGRATION...")
		if err = mig.migrate(); err != nil {
			return err
		}
		logrus.Infoln("MIGRATE OK")
	}

	return nil
}

func (mig *Migrator) normalUpdate() (err error) {
	mig.LocalScripts.MarkPending(mig.DB())

	if err = mig.patchBeforeUpdating(); err != nil {
		return errors.Wrapf(err, "failed to patch before this time updating")
	}

	// Erda mysql lint
	if !mig.SkipMigrationLint() {
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
	if err = mig.LocalScripts.InstalledChangesLint(mig.DB()); err != nil {
		return err
	}
	logrus.Infoln("INSTALLED CHANGES LINT OK")

	if !mig.SkipSandbox() {
		// copy database snapshot to sandbox
		logrus.Infoln("COPY CURRENT DATABASE STRUCTURE TO SANDBOX....")
		if err = mig.snap.RecoverTo(mig.SandBox()); err != nil {
			return err
		}
		logrus.Infoln("COPY CURRENT DATABASE STRUCTURE TO SANDBOX OK")

		// migrate in sandbox
		logrus.Infoln("DO MIGRATION IN SANDBOX....")
		if err = mig.migrateSandbox(); err != nil {
			return err
		}
		logrus.Infoln("MIGRATE IN SANDBOX OK")
	}

	// pre migrate data
	if !mig.SkipPreMigrate() && !mig.SkipMigrate() {
		logrus.Infoln("DO PRE-MIGRATION....")
		if err = mig.preMigrate(); err != nil {
			return err
		}
		logrus.Infoln("PRE-MIGRATE OK")
	}

	// migrate data
	if !mig.SkipMigrate() {
		logrus.Infoln("DO MIGRATION....")
		if err = mig.migrate(); err != nil {
			return err
		}
		logrus.Infoln("MIGRATE OK")
	}

	return nil
}

func (mig *Migrator) firstTimeUpdate() (err error) {
	// the correct state in this time is db.Schema == baseScriptsSchema,
	// if not, returns error, manual intervention.
	// marks all baseScripts !pending, marks others pending

	if err = mig.patchBeforeUpdating(); err != nil {
		return errors.Wrapf(err, "failed to patch before first time updating")
	}

	// compare local base schema and cloud base schema for every service
	logrus.Infoln("COMPARE LOCAL SCHEMA AND CLOUD SCHEMA FOR EVERY SERVICE... ..")
	for _, service := range mig.LocalScripts.Services {
		if equal := service.BaselineEqualCloud(mig.DB()); !equal.Equal() {
			return errors.Errorf("local base schema is not equal with cloud schema: %s", equal.Reason())
		}
	}

	// record base
	logrus.Infoln("RECORD BASE... ..")
	new(HistoryModel).CreateTable(mig.DB())
	now := time.Now()
	for moduleName, module := range mig.LocalScripts.Services {
		for i := range module.Scripts {
			if module.Scripts[i].IsBaseline() {
				module.Scripts[i].Pending = false
				record := HistoryModel{
					ID:           0,
					CreatedAt:    now,
					UpdatedAt:    now,
					ServiceName:  moduleName,
					Filename:     filepath.Base(module.Scripts[i].GetName()),
					Checksum:     module.Scripts[i].Checksum(),
					InstalledBy:  "",
					InstalledOn:  "",
					LanguageType: string(module.Scripts[i].Type),
					Reversed:     ddlreverser.ReverseCreateTableStmts(module.Scripts[i]),
				}
				if err := mig.DB().Create(&record).Error; err != nil {
					return err
				}
			}
		}
	}

	return mig.normalUpdate()
}

func (mig *Migrator) reverse(reversing []string, reverseSlice bool) error {
	if len(reversing) == 0 {
		return nil
	}
	if reverseSlice {
		ddlreverser.ReverseSlice(reversing)
	}
	for _, s := range mig.reversing {
		if err := mig.DB().Exec(s).Error; err != nil {
			return errors.Errorf("failed to exec reversing SQL: %s", s)
		}
	}
	return nil
}

func (mig *Migrator) migrateSandbox() (err error) {
	if err = mig.destructiveLint(); err != nil {
		return err
	}

	// install every module
	for moduleName, module := range mig.LocalScripts.Services {
		for _, script := range module.Scripts {
			if !script.Pending {
				continue
			}
			switch script.Type {
			case ScriptTypeSQL:
				after := func(tx *gorm.DB, err error) {
					tx.Commit()
				}
				if err := mig.installSQL(script, mig.SandBox, mig.SandBox().Begin, after); err != nil {
					return errors.Wrapf(err, "failed to migrate in sandbox: %+v",
						map[string]interface{}{"moduleName": moduleName, "filename": script.GetName(), "type": ScriptTypeSQL})
				}
			case ScriptTypePython:
				if err := mig.installPy(script, module, mig.sandboxSettings, true); err != nil {
					return errors.Wrapf(err, "failed to migrate in sandbox: %+v",
						map[string]interface{}{"moduleName": moduleName, "filename": script.GetName(), "type": ScriptTypePython})
				}
			}
		}
	}

	return err
}

// pre migrate data SQLs, all applied in this runtime will be rollback
func (mig *Migrator) preMigrate() (err error) {
	if err = mig.destructiveLint(); err != nil {
		return err
	}

	// finally roll all DDL back
	defer func() {
		logrus.Infoln("	REVERSE PRE-MIGRATIONS: all schema migration")
		if err := mig.reverse(mig.reversing, true); err != nil {
			logrus.Fatalln(err)
		}
	}()

	mig.reversing = nil

	// install every module
	for moduleName, module := range mig.LocalScripts.Services {
		for _, script := range module.Scripts {
			if !script.Pending {
				continue
			}

			switch script.Type {
			case ScriptTypeSQL:
				after := func(tx *gorm.DB, err error) {
					logrus.WithField("module name", moduleName).
						WithField("script name", script.GetName()).
						Infoln("	ROLLBACK PRE-MIGRATIONS: current script data migration")
					tx.Rollback()
				}
				if err := mig.installSQL(script, mig.DB, mig.DB().Begin, after); err != nil {
					return errors.Wrapf(err, "failed to pre-migrate: %+v",
						map[string]interface{}{"module name": moduleName, "script name": script.GetName(), "type": ScriptTypeSQL})
				}
			case ScriptTypePython:
				if err := mig.installPy(script, module, mig.dbSettings, false); err != nil {
					return errors.Wrapf(err, "failed to pre-migrate: %+v",
						map[string]interface{}{"module name": moduleName, "script name": script.GetName(), "type": ScriptTypePython})
				}
			}
		}
	}

	return nil
}

func (mig *Migrator) migrate() error {
	now := time.Now()

	// install every service
	for moduleName, module := range mig.LocalScripts.Services {
		logrus.WithField("module", moduleName).Infoln()
		for _, script := range module.Scripts {
			if !script.Pending {
				continue
			}

			logrus.WithField("script name", script.GetName()).Infoln("install")
			switch script.Type {
			case ScriptTypeSQL:
				after := func(tx *gorm.DB, err error) {
					if err != nil {
						tx.Rollback()
						mig.reverse(script.Reversing, true)
					} else {
						tx.Commit()
					}
				}
				if err := mig.installSQL(script, mig.DB, mig.DB().Begin, after); err != nil {
					return errors.Wrapf(err, "failed to migrate: %+v",
						map[string]interface{}{"module name": moduleName, "script name": script.GetName(), "type": ScriptTypeSQL})
				}

			case ScriptTypePython:
				if err := mig.installPy(script, module, mig.dbSettings, true); err != nil {
					return errors.Wrapf(err, "failed to migrate: %+v",
						map[string]interface{}{"module name": moduleName, "script name": script.GetName(), "type": ScriptTypePython})
				}
			}

			// record it
			ddlreverser.ReverseSlice(script.Reversing)
			record := HistoryModel{
				ID:           0,
				CreatedAt:    now,
				UpdatedAt:    now,
				ServiceName:  moduleName,
				Filename:     script.GetName(),
				Checksum:     script.Checksum(),
				InstalledBy:  "",
				InstalledOn:  "",
				LanguageType: string(script.Type),
				Reversed:     strings.Join(script.Reversing, "\n"),
			}
			if err := mig.DB().Create(&record).Error; err != nil {
				return errors.Wrapf(err, "internal error: failed to record migration")
			}
		}
	}

	return nil
}

func (mig *Migrator) patchBeforeUpdating() error {
	now := time.Now()
	module := mig.LocalScripts.GetPatches()
	if module == nil {
		return nil
	}

	logrus.WithField("module", module.Name)
	for _, script := range module.Scripts {
		if tx := mig.DB().Where(map[string]interface{}{"filename": script.GetName()}); tx.RowsAffected > 0 {
			continue
		}

		logrus.WithField("script name", script.GetName()).Infoln("patch it")
		switch script.Type {
		case ScriptTypeSQL:
			after := func(tx *gorm.DB, err error) {
				tx.Commit()
			}
			if err := mig.installSQL(script, mig.DB, mig.DB().Begin, after); err != nil {
				return errors.Wrapf(err, "failed to patch, module name: %s, script name: %s, type: %s",
					module.Name, script.GetName(), ScriptTypeSQL)
			}

		case ScriptTypePython:
			if err := mig.installPy(script, module, mig.dbSettings, true); err != nil {
				return errors.Wrapf(err, "failed to patch, module name: %s, script name: %s, type: %s",
					module.Name, script.GetName(), ScriptTypePython)
			}
		}

		// record it
		record := HistoryModel{
			ID:           0,
			CreatedAt:    now,
			UpdatedAt:    now,
			ServiceName:  module.Name,
			Filename:     script.GetName(),
			Checksum:     script.Checksum(),
			InstalledBy:  "",
			InstalledOn:  "",
			LanguageType: string(script.Type),
			Reversed:     "",
		}
		if err := mig.DB().Create(&record).Error; err != nil {
			return errors.Wrapf(err, "patch error: failed to record migration")
		}
	}

	return nil
}

func (mig *Migrator) installSQL(s *Script, exec func() *gorm.DB, begin func(opts ...*sql.TxOptions) *gorm.DB, after func(tx *gorm.DB, err error)) (err error) {
	tx := begin()
	defer after(tx, err)
	defer func() {
		mig.reversing = append(mig.reversing, s.Reversing...)
	}()

	s.Reversing = nil

	for _, node := range s.DDLNodes() {
		var (
			reverse string
			ok      bool
		)
		reverse, ok, err = ddlreverser.ReverseDDLWithSnapshot(exec(), node)
		if err != nil {
			return errors.Wrapf(err, "failed to generate reversed DDL: %+v",
				map[string]string{"scritpName": s.GetName(), "SQL": node.Text()})
		}
		if ok {
			s.Reversing = append(s.Reversing, reverse)
		}

		if err = exec().Exec(node.Text()).Error; err != nil {
			return errors.Wrapf(err, "failed to do schema migration: %+v",
				map[string]string{"scriptName": s.GetName(), "SQL": node.Text()})
		}
	}

	for _, node := range s.DMLNodes() {
		if err = tx.Exec(node.Text()).Error; err != nil {
			return errors.Wrapf(err, "failed to do data migration: %+v",
				map[string]string{"scriptName": s.GetName(), "SQL": node.Text()})
		}
	}

	return nil
}

func (mig *Migrator) installPy(s *Script, module *Module, settings *pygrator.Settings, commit bool) error {
	var p = pygrator.Package{
		DeveloperScript: s,
		Requirements:    module.PythonRequirementsText,
		Settings:        *settings,
		Commit:          commit,
	}
	if len(p.Requirements) == 0 {
		p.Requirements = []byte(pygrator.BaseRequirements)
	}

	if err := p.Make(); err != nil {
		return err
	}
	defer p.Remove()

	return p.Run()
}

func (mig *Migrator) destructiveLint() error {
	text, ok := mig.LocalScripts.HasDestructiveOperationInPending()
	if ok {
		return errors.Errorf("there is desctructive SQL in pending scripts, SQL: %s", text)
	}

	return nil
}
