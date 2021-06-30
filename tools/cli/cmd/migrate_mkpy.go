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

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/database/pyorm/pattern"
	"github.com/erda-project/erda/pkg/database/sqlparser/migrator"
	"github.com/erda-project/erda/tools/cli/command"
)

var MigratePy = command.Command{
	ParentName:     "Migrate",
	Name:           "mkpy",
	ShortHelp:      "make a python migration script pattern",
	LongHelp:       "make a python migration scritp pattern.",
	Example:        "erda-cli migrate py --module my_module --name my_script_name",
	Hidden:         false,
	DontHideCursor: false,
	Args:           nil,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "",
			Name:         "workdir",
			Doc:          "workdir",
			DefaultValue: ".",
		},
		command.StringFlag{
			Short:        "m",
			Name:         "module",
			Doc:          "migration module name",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "n",
			Name:         "name",
			Doc:          "script name",
			DefaultValue: "",
		},
		command.StringListFlag{
			Short:        "",
			Name:         "tables",
			Doc:          "dependency tables",
			DefaultValue: nil,
		},
	},
	Run: RunMigratePy,
}

func RunMigratePy(ctx *command.Context, workdir, module, name string, tables []string) error {
	moduleInfos, err := ioutil.ReadDir(workdir)
	if err != nil {
		return err
	}

	var (
		validModule = false
		scripts     = make(migrator.Module, 0)
	)

	for _, moduleInfo := range moduleInfos {
		if !moduleInfo.IsDir() {
			continue
		}
		if moduleInfo.Name() == module {
			validModule = true
		}
		fileInfos, err := ioutil.ReadDir(moduleInfo.Name())
		if err != nil {
			return errors.Wrap(err, "failed to ReadDir, directory name: %s")
		}
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}
			if ext := filepath.Ext(fileInfo.Name()); strings.EqualFold(ext, string(migrator.ScriptTypeSQL)) {
				script, err := migrator.NewScript(workdir, filepath.Join(workdir, moduleInfo.Name(), fileInfo.Name()))
				if err != nil {
					return errors.Wrapf(err, "failed to NewScript")
				}
				scripts = append(scripts, script)
			}
		}
	}
	if !validModule {
		return errors.Errorf("invalid module name: %s", module)
	}

	if err = genRequirements(filepath.Join(workdir, module)); err != nil {
		return errors.Wrap(err, "failed to generate requirements.txt")
	}

	scripts.Sort()

	var (
		schema  = scripts.Schema()
		script  pattern.DeveloperScript
		tablesM = make(map[string]bool)
	)
	for _, tableName := range tables {
		tablesM[tableName] = true
	}
	for _, definition := range schema.TableDefinitions {
		// if user specifies the tables, skip the table which is not in the tables
		if definition.CreateStmt == nil {
			continue
		}
		if _, ok := tablesM[definition.CreateStmt.Table.Name.String()]; !ok {
			continue
		}

		// make table definition to Django Model
		model, err := pattern.CreateTableStmtToModel(definition.CreateStmt)
		if err != nil {
			return errors.Wrapf(err, "failed to CreateTableStmtToModel")
		}
		var buf = bytes.NewBuffer(nil)
		if err = pattern.GenModel(buf, *model); err != nil {
			return errors.Wrapf(err, "failed to GenModel")
		}
		script.Models = append(script.Models, buf.String())
	}

	name = strings.TrimSuffix(name, filepath.Ext(name)) + string(migrator.ScriptTypePython)
	var filename = filepath.Join(workdir, module, name)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to OpenFile: %s", filename)
	}
	defer f.Close()
	if err = pattern.GenDeveloperScript(f, script); err != nil {
		return err
	}

	return nil
}

func genRequirements(dir string) error {
	filename := filepath.Join(dir, pattern.RequirementsFilename)
	_, err := ioutil.ReadFile(filename)
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *os.PathError:
		fmt.Printf("%T", err.(*os.PathError).Err)
	}
	if os.IsNotExist(err) {
		if err := ioutil.WriteFile(filename, []byte(pattern.Requirements), 0644); err != nil {
			return err
		}
	}
	return err
}
