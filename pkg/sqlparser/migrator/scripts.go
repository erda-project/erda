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

	"github.com/erda-project/erda/pkg/sqllint"
	"github.com/erda-project/erda/pkg/sqllint/rules"
)

// Scripts is the set of Module
type Scripts struct {
	Workdir       string
	Dirname       string
	ServicesNames []string
	Services      map[string]Module

	rulers      []rules.Ruler
	markPending bool
	destructive int
}

// NewScripts range the directory
func NewScripts(parameters Parameters) (*Scripts, error) {
	var (
		modulesNames []string
		services     = make(map[string]Module, 0)
	)
	dirname := filepath.Join(parameters.Workdir(), parameters.MigrationDir())
	modulesInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ReadDir %s", dirname)
	}
	var (
		moduleList = parameters.Modules()
		modules    = make(map[string]bool, len(moduleList))
	)
	for _, module := range moduleList {
		modules[module] = true
	}
	for _, moduleInfo := range modulesInfos {
		if !moduleInfo.IsDir() {
			continue
		}
		// specified modules and this service is in specified modules then to continue
		if _, ok := modules[moduleInfo.Name()]; len(modules) > 0 && !ok {
			continue
		}

		modulesNames = append(modulesNames, moduleInfo.Name())
		scripts := make(Module, 0)

		dirname := filepath.Join(parameters.Workdir(), parameters.MigrationDir(), moduleInfo.Name())
		serviceDirInfos, err := ioutil.ReadDir(dirname)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to ReadDir %s", dirname)
		}

		for _, fileInfo := range serviceDirInfos {
			if !fileInfo.IsDir() && strings.EqualFold(filepath.Ext(fileInfo.Name()), ".sql") {
				script, err := NewScript(parameters.Workdir(), filepath.Join(parameters.MigrationDir(), moduleInfo.Name(), fileInfo.Name()))
				if err != nil {
					return nil, errors.Wrap(err, "failed to NewScript")
				}
				scripts = append(scripts, script)
			}
		}

		scripts.Sort()
		services[moduleInfo.Name()] = scripts
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
	scripts, ok := s.Services[serviceName]
	return scripts, ok
}

func (s *Scripts) Lint() error {
	if !s.markPending {
		s.MarkPending(DB())
	}

	linter := sqllint.New(s.rulers...)
	for serviceName, scripts := range s.Services {
		for _, script_ := range scripts {
			if !script_.isBase && script_.Pending {
				if err := linter.Input(script_.Rawtext, filepath.Join(s.Dirname, serviceName, script_.Name)); err != nil {
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
	for serviceName, service := range s.Services {
		tableNames := make(map[string]bool, 0)
		for _, script_ := range service {
			for _, ddl := range script_.DDLNodes() {
				switch ddl.(type) {
				case *ast.CreateTableStmt:
					tableName := ddl.(*ast.CreateTableStmt).Table.Name.String()
					tableNames[tableName] = true
				case *ast.AlterTableStmt:
					tableName := ddl.(*ast.AlterTableStmt).Table.Name.String()
					if _, ok := tableNames[tableName]; !ok {
						return errors.Errorf("the table tried to alter not exists, may it not created in this service directory. filename: %s, text:\n%s",
							filepath.Join(s.Dirname, serviceName, script_.Name), ddl.Text())
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
	for serviceName, service := range s.Services {
		for i := range service {
			var record HistoryModel
			if tx := tx.Where(map[string]interface{}{
				"service_name": serviceName,
				"filename":     filepath.Base(service[i].Name),
			}).
				First(&record); tx.Error != nil || tx.RowsAffected == 0 {
				service[i].Pending = true
			} else {
				service[i].Pending = false
				service[i].Record = &record
			}
		}
	}

	s.markPending = true
}

func (s *Scripts) InstalledChangesLint() error {
	if !s.markPending {
		s.MarkPending(DB())
	}

	for serviceName, scripts_ := range s.Services {
		var (
			pending     bool
			pendingName string
		)
		for _, script_ := range scripts_ {
			if pending {
				if script_.Pending {
					continue
				}
				return errors.Errorf("some uninstalled script is before a installed script. The service name: %s, pending filename: %s, the installed filename: %s",
					serviceName, pendingName, script_.Name)
			}
			if script_.Pending {
				pending = true
				pendingName = script_.Name
				continue
			}
			if script_.Checksum() != script_.Record.Checksum {
				return errors.Errorf("the installed script is changed in local. The service name: %s, script filename: %s",
					serviceName, script_.Name)
			}
		}
	}
	return nil
}

// SameNameLint lint whether there is same script name in different directories
func (s *Scripts) SameNameLint() error {
	var m = make(map[string]string)
	for _, scripts_ := range s.Services {
		for _, script_ := range scripts_ {
			if existsName, ok := m[filepath.Base(script_.Name)]; ok {
				return errors.Errorf("there is not allowed same script name in different directory: %s:%s",
					existsName, script_.Name)
			} else {
				m[filepath.Base(script_.Name)] = script_.Name
			}
		}
	}
	return nil
}

func (s *Scripts) HasDestructiveOperationInPending() bool {
	if s.destructive == 1 {
		return true
	}
	if s.destructive == -1 {
		return false
	}

	s.destructive = -1
	for _, scripts := range s.Services {
		for _, script := range scripts {
			if !script.Pending {
				continue
			}
			for _, node := range script.Nodes {
				switch node.(type) {
				case *ast.DropDatabaseStmt, *ast.DropTableStmt, *ast.TruncateTableStmt:
					s.destructive = 1
				case *ast.AlterTableStmt:
					for _, spec := range node.(*ast.AlterTableStmt).Specs {
						switch spec.Tp {
						case ast.AlterTableDropColumn:
							s.destructive = 1
						}
					}
				}
			}
		}
	}

	return s.destructive == 1
}
