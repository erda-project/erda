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

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/configuration"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqlparser/migrator"
	"github.com/erda-project/erda/tools/cli/command"
)

const (
	baseScriptLabel  = "# MIGRATION_BASE"
	baseScriptLabel2 = "-- MIGRATION_BASE"
	baseScriptLabel3 = "/* MIGRATION_BASE */"
)

var MigrationLint = command.Command{
	ParentName:     "Migrate",
	Name:           "lint",
	ShortHelp:      "Erda MySQL Migration lint",
	LongHelp:       "Erda MySQL Migration lint",
	Example:        "erda-cli migrate lint --input=. --config=default.yaml --detail",
	Hidden:         false,
	DontHideCursor: false,
	Args:           nil,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "",
			Name:         "input",
			Doc:          "[optional] the file or directory for linting",
			DefaultValue: ".",
		},
		command.StringFlag{
			Short:        "",
			Name:         "lint-config",
			Doc:          "[optional] the lint config file",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "",
			Name:         "no-detail",
			Doc:          "[optional] do not print details of lint result",
			DefaultValue: false,
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          "[optional] result output file name",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "",
			Name:         "custom",
			Doc:          "custom directory",
			DefaultValue: false,
		},
	},
	Run: RunMigrateLint,
}

func RunMigrateLint(ctx *command.Context, input, config string, noDetail bool, output string, custom bool) (err error) {
	log.Printf("Erda MySQL Lint the input file or directory: %s", input)
	if custom {
		return CustomMigrateLint(ctx, input, config, noDetail, output)
	}
	return StandardMigrateLint(ctx, input, config)
}

func StandardMigrateLint(ctx *command.Context, input, config string) (err error) {
	defer func() {
		if err != nil {
			log.Fatalln(err)
		}
	}()

	var p = scriptsParameters{
		migrationDir: input,
		rules:        nil,
	}

	rulers, err := loadRules(config)
	if err != nil {
		return err
	}
	p.rules = rulers

	scripts, err := migrator.NewScripts(&p)
	if err != nil {
		return err
	}
	scripts.IgnoreMarkPending()

	if err = scripts.SameNameLint(); err != nil {
		return err
	}

	if err = scripts.AlterPermissionLint(); err != nil {
		return err
	}

	if err = scripts.Lint(); err != nil {
		return err
	}

	log.Println("Erda MySQL Migration Lint OK")
	return nil
}

func CustomMigrateLint(ctx *command.Context, input, config string, noDetail bool, output string) (err error) {
	exitFunc := func(_ int) {
		if err != nil {
			os.Exit(1)
		}
	}
	defer exitFunc(1)

	files := new(walk).walk(input, ".sql").filenames()

	rulers, err := loadRules(config)
	if err != nil {
		return err
	}

	linter := sqllint.New(rulers...)

	for _, filename := range files {
		var data []byte
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "failed to open file, filename: %s", filename)
		}
		if isBaseScript(data) {
			continue
		}
		if err = linter.Input(data, filename); err != nil {
			return errors.Wrapf(err, "failed to run Erda MySQL Lint on the SQL script, filename: %s", filename)
		}
	}

	if len(linter.Errors()) == 0 {
		log.Println("Erda MySQL Migration Lint OK")
		return nil
	}
	exitFunc = os.Exit

	var out = log.Writer()
	if output != "" {
		var f *os.File
		f, err = os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("failed to OpenFile, filename: %s", output)
		} else {
			defer f.Close()
			out = f
		}
	}

	if _, err = fmt.Fprintln(out, linter.Report()); err != nil {
		return errors.Wrapf(err, "failed to print lint report")
	}

	if !noDetail {
		for src, errs := range linter.Errors() {
			if _, err = fmt.Fprintln(out, src); err != nil {
				return errors.Wrapf(err, "failed to print lint error")
			}
			for _, e := range errs {
				if _, err = fmt.Fprintln(out, e); err != nil {
					return errors.Wrapf(err, "failed to print lint error")
				}
			}
		}
	}

	return errors.New("some errors in your migrations")
}

type walk struct {
	files []string
}

func (w *walk) filenames() []string {
	return w.files
}

func (w *walk) walk(input, suffix string) *walk {
	infos, err := ioutil.ReadDir(input)
	if err != nil {
		w.files = append(w.files, input)
		return w
	}

	for _, info := range infos {
		if info.IsDir() {
			w.walk(filepath.Join(input, info.Name()), suffix)
			continue
		}
		if strings.ToLower(path.Ext(info.Name())) == strings.ToLower(suffix) {
			file := filepath.Join(input, info.Name())
			w.files = append(w.files, file)
		}
	}

	return w
}

type scriptsParameters struct {
	migrationDir string
	rules        []rules.Ruler
}

func (p *scriptsParameters) Workdir() string {
	return "."
}

func (p *scriptsParameters) MigrationDir() string {
	return p.migrationDir
}

func (p *scriptsParameters) Modules() []string {
	return nil
}

func (p *scriptsParameters) Rules() []rules.Ruler {
	return p.rules
}

func isBaseScript(data []byte) bool {
	return bytes.HasPrefix(data, []byte(baseScriptLabel)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel2)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel3))
}

func loadRules(config string) (rulers []rules.Ruler, err error) {
	var lintCfg *configuration.Configuration
	rulers = configuration.DefaultRulers()

	if config != "" {
		lintCfg, err = configuration.FromLocal(config)
		if err != nil {
			return nil, errors.Wrapf(err, "filed to read lint configuration from local")
		}
		rulers, err = lintCfg.Rulers()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate lint rulers from lint configuration")
		}
	}

	return rulers, nil
}
