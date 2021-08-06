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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqlparser/pygrator"
)

const (
	patchesModuleName = ".patches"
	patchInit         = "patch.sql"
)

type ScriptsParameters interface {
	// Workdir gets pipeline node workdir
	Workdir() string

	// MigrationDir gets migration scripts direction from repo, like .dice/migrations or 4.1/sqls
	MigrationDir() string

	// Modules is the modules for installing.
	// if is nil, to install all modules in the MigrationDir()
	Modules() []string

	Rules() []rules.Ruler
}

// Scripts is the set of Module
type Scripts struct {
	Workdir       string
	Dirname       string
	ServicesNames []string
	Services      map[string]*Module
	Patches       *Module

	rulers          []rules.Ruler
	markPending     bool
	destructive     int
	destructiveText string
}

// NewScripts range the directory
func NewScripts(parameters ScriptsParameters) (*Scripts, error) {
	var (
		modulesNames []string
		services     = make(map[string]*Module, 0)
	)
	dirname := filepath.Join(parameters.Workdir(), parameters.MigrationDir())
	modulesInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ReadDir %s", dirname)
	}
	var (
		moduleList = parameters.Modules()
		modules    = map[string]bool{patchesModuleName: true}
	)
	for _, moduleName := range moduleList {
		if moduleName != "" {
			modules[moduleName] = true
		}
	}
	for _, moduleInfo := range modulesInfos {
		if !moduleInfo.IsDir() {
			continue
		}
		// specified modules and this service is in specified modules then to continue
		if _, ok := modules[moduleInfo.Name()]; len(moduleList) > 0 && !ok {
			continue
		}

		var module Module
		module.Name = moduleInfo.Name()
		modulesNames = append(modulesNames, moduleInfo.Name())

		dirname := filepath.Join(parameters.Workdir(), parameters.MigrationDir(), moduleInfo.Name())
		serviceDirInfos, err := ioutil.ReadDir(dirname)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to ReadDir %s", dirname)
		}

		for _, fileInfo := range serviceDirInfos {
			if fileInfo.IsDir() {
				continue
			}

			// read requirements.txt
			if strings.EqualFold(fileInfo.Name(), pygrator.RequirementsFilename) {
				module.PythonRequirementsText, err = ioutil.ReadFile(filepath.Join(parameters.Workdir(), parameters.MigrationDir(), moduleInfo.Name(), fileInfo.Name()))
				if err != nil {
					return nil, err
				}
			}

			// read script (.sql or .py)
			if ext := filepath.Ext(fileInfo.Name()); strings.EqualFold(ext, string(ScriptTypeSQL)) || strings.EqualFold(ext, string(ScriptTypePython)) {
				script, err := NewScript(parameters.Workdir(), filepath.Join(parameters.MigrationDir(), moduleInfo.Name(), fileInfo.Name()))
				if err != nil {
					return nil, errors.Wrap(err, "failed to NewScript")
				}
				module.Scripts = append(module.Scripts, script)
			}
		}

		module.Sort()
		services[moduleInfo.Name()] = &module
	}

	var scritps = &Scripts{
		Workdir:       parameters.Workdir(),
		Dirname:       parameters.MigrationDir(),
		ServicesNames: modulesNames,
		Services:      services,
		rulers:        parameters.Rules(),
		markPending:   false,
		destructive:   0,
	}
	if module, ok := scritps.Services[patchesModuleName]; ok {
		scritps.Patches = module
		delete(scritps.Services, patchesModuleName)
	} else {
		scritps.Patches = new(Module)
	}

	return scritps, nil
}

func (s *Scripts) Get(serviceName string) ([]*Script, bool) {
	module, ok := s.Services[serviceName]
	return module.Scripts, ok
}

func (s *Scripts) GetPatches() *Module {
	return s.Patches
}

func (s *Scripts) GetScriptByFilename(filename string) (*Module, *Script, bool) {
	for _, mod := range s.Services {
		script, ok := mod.GetScriptByFilename(filename)
		if ok {
			return mod, script, true
		}
	}
	return nil, nil, false
}

func (s *Scripts) Lint() error {
	if !s.markPending {
		return errors.New("scripts did not mark if is pending, please mark it and then do Lint")
	}

	linter := sqllint.New(s.rulers...)
	for moduleName, module := range s.Services {
		for _, script := range module.Scripts {
			if !script.isBase && script.Pending && script.Type == ScriptTypeSQL {
				if err := linter.Input(script.Rawtext, filepath.Join(s.Dirname, moduleName, script.GetName())); err != nil {
					return err
				}
			}
		}
	}
	if len(linter.Errors()) == 0 {
		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, linter.Report())
	for src, es := range linter.Errors() {
		logrus.Println(src)
		for _, err := range es {
			_, _ = fmt.Fprintln(os.Stdout, err)
		}
		_, _ = fmt.Fprintln(os.Stdout)
	}

	return errors.New("many lint errors")
}

func (s *Scripts) AlterPermissionLint() error {
	for moduleName, module := range s.Services {
		if moduleName == patchesModuleName {
			continue
		}

		tableNames := make(map[string]bool, 0)
		for _, script := range module.Scripts {
			for _, ddl := range script.DDLNodes() {
				switch ddl.(type) {
				case *ast.CreateTableStmt:
					tableName := ddl.(*ast.CreateTableStmt).Table.Name.String()
					tableNames[tableName] = true
				case *ast.AlterTableStmt:
					tableName := ddl.(*ast.AlterTableStmt).Table.Name.String()
					if _, ok := tableNames[tableName]; !ok {
						return errors.Errorf("the table you tried to alter is not exists, may it not created in this module directory. filename: %s, text:\n%s",
							filepath.Join(s.Dirname, moduleName, script.GetName()), ddl.Text())
					}
				default:
					continue
				}
			}
		}
	}
	return nil
}

func (s *Scripts) MarkPending(tx *gorm.DB) {
	for _, module := range s.Services {
		for i := range module.Scripts {
			var record HistoryModel
			if tx := tx.Where(map[string]interface{}{"filename": module.Scripts[i].GetName()}).
				First(&record); tx.Error != nil || tx.RowsAffected == 0 {
				module.Scripts[i].Pending = true
			} else {
				module.Scripts[i].Pending = false
				module.Scripts[i].Record = &record
			}
		}
	}

	s.markPending = true
}

func (s *Scripts) IgnoreMarkPending() {
	s.markPending = true
}

func (s *Scripts) InstalledChangesLint(ctx *context.Context, db *gorm.DB) error {
	var (
		patchesList        []string
		missingPatchesList []string
	)
	for moduleName, module := range s.Services {
		for _, script := range module.Scripts {
			if script.Record == nil {
				script.Record = new(HistoryModel)
			}
			db := db.Where(map[string]interface{}{"filename": script.GetName()}).First(script.Record)
			if db.Error != nil {
				continue
			}
			if script.Checksum() != script.Record.Checksum {
				logrus.Warnf("the installed file is changed on local, filename: %s, expected checksum: %s, actual checksum: %s",
					script.GetName(), script.Checksum(), script.Record.Checksum)
				filename := patchPrefix + script.GetName()
				if _, ok := s.Patches.GetScriptByFilename(filename); ok {
					logrus.Infof("found patch file and append it to the list, filename: %s", filename)
					patchesList = append(patchesList, filename)
				} else {
					logrus.Errorf("missing path file, filename: %s", filename)
					missingPatchesList = append(missingPatchesList, filepath.Join(moduleName, script.GetName()))
				}
			}
		}
	}

	if len(missingPatchesList) > 0 {
		return errors.Errorf("these installed script is changed on local and mising paches: %s", strings.Join(missingPatchesList, ","))
	}

	*ctx = context.WithValue(*ctx, patchesKey, patchesList)
	return nil
}

// SameNameLint lint whether there is same script name in different directories
func (s *Scripts) SameNameLint() error {
	// m's key is script file name, value is module name
	var m = make(map[string]string)
	for curModuleName, module := range s.Services {
		for _, script := range module.Scripts {
			if moduleName, ok := m[script.GetName()]; ok {
				return errors.Errorf("not allowed same script name in different directory, filename: %s, modules: %s, %s",
					script.GetName(), curModuleName, moduleName)
			} else {
				m[script.GetName()] = curModuleName
			}
		}
	}
	return nil
}

func (s *Scripts) HasDestructiveOperationInPending() (string, bool) {
	if s.destructive == 1 {
		return s.destructiveText, true
	}
	if s.destructive == -1 {
		return "", false
	}

	s.destructive = -1
	for _, module := range s.Services {
		for _, script := range module.Scripts {
			if !script.Pending || script.IsBaseline() {
				continue
			}
			for _, node := range script.Nodes {
				switch node.(type) {
				case *ast.DropDatabaseStmt, *ast.DropTableStmt, *ast.TruncateTableStmt:
					s.destructive = 1
					s.destructiveText = node.Text()
					return s.destructiveText, true
				case *ast.AlterTableStmt:
					for _, spec := range node.(*ast.AlterTableStmt).Specs {
						switch spec.Tp {
						case ast.AlterTableDropColumn:
							s.destructive = 1
							s.destructiveText = node.Text()
							return s.destructiveText, true
						}
					}
				}
			}
		}
	}

	return "", false
}

func (s *Scripts) FreshBaselineModules(db *gorm.DB) map[string]*Module {
	var modules = make(map[string]*Module)
	for name, mod := range s.Services {
		mod := mod.FilterFreshBaseline(db)
		if len(mod.Scripts) > 0 {
			modules[name] = mod
		}
	}
	return modules
}
