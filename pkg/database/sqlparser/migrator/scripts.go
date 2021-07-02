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
	pygrator2 "github.com/erda-project/erda/pkg/database/sqlparser/pygrator"
)

// Scripts is the set of Module
type Scripts struct {
	Workdir       string
	Dirname       string
	ServicesNames []string
	Services      map[string]*Module

	rulers      []rules.Ruler
	markPending bool
	destructive int
	destructiveText string
}

// NewScripts range the directory
func NewScripts(parameters Parameters) (*Scripts, error) {
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
		modules    = make(map[string]bool)
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
		if _, ok := modules[moduleInfo.Name()]; len(modules) > 0 && !ok {
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
			if strings.EqualFold(fileInfo.Name(), pygrator2.RequirementsFilename) {
				module.PythonRequirementsText, err = ioutil.ReadFile(filepath.Join(parameters.Workdir(), parameters.MigrationDir(), moduleInfo.Name(), fileInfo.Name()))
				if err != nil {
					return nil, err
				}
			}

			// read script (.sql or .py)
			if ext := filepath.Ext(fileInfo.Name());
			strings.EqualFold(ext, string(ScriptTypeSQL)) ||			strings.EqualFold(ext, string(ScriptTypePython)) {
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

	return &Scripts{
		Workdir:       parameters.Workdir(),
		Dirname:       parameters.MigrationDir(),
		ServicesNames: modulesNames,
		Services:      services,
		rulers:        parameters.Rules(),
		markPending:   false,
		destructive:   0,
	}, nil
}

func (s *Scripts) Get(serviceName string) ([]*Script, bool) {
	module, ok := s.Services[serviceName]
	return module.Scripts, ok
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
						return errors.Errorf("the table tried to alter not exists, may it not created in this module directory. filename: %s, text:\n%s",
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
	for moduleName, module := range s.Services {
		for i := range module.Scripts {
			var record HistoryModel
			if tx := tx.Where(map[string]interface{}{
				"service_name": moduleName,
				"filename":     module.Scripts[i].GetName(),
			}).
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

func (s *Scripts) InstalledChangesLint() error {
	if !s.markPending {
		return errors.New("scripts did not mark if is pending, please mark it and then do InstalledChangesLint")
	}

	for moduleName, module := range s.Services {
		var (
			pending     bool
			pendingName string
		)
		for _, script := range module.Scripts {
			switch {
			case pending && script.Pending:
				continue
			case pending:
				return errors.Errorf("some uninstalled script is ranked before a installed script. The service name: %s, pending filename: %s, the installed filename: %s",
					moduleName, pendingName, script.GetName())
			case script.Pending:
				pending = true
				pendingName = script.GetName()
				continue
			}

			if script.Checksum() != script.Record.Checksum {
				return errors.Errorf("the installed script is changed in local. The service name: %s, script filename: %s",
					moduleName, script.GetName())
			}
		}
	}
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
			if !script.Pending {
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
