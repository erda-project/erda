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
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/configuration"
	"github.com/erda-project/erda/tools/cli/command"
)

const (
	baseScriptLabel  = "# MIGRATION_BASE"
	baseScriptLabel2 = "-- MIGRATION_BASE"
	baseScriptLabel3 = "/* MIGRATION_BASE */"
)

var MigrationLint = command.Command{
	ParentName:     "",
	Name:           "miglint",
	ShortHelp:      "Erda MySQL Migration lint",
	LongHelp:       "Erda MySQL Migration lint",
	Example:        "erda-cli miglint --input=. config=default.yaml --detail",
	Hidden:         false,
	DontHideCursor: false,
	Args:           nil,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "f",
			Name:         "filename",
			Doc:          "[optional] the file or directory for linting",
			DefaultValue: ".",
		},
		command.StringFlag{
			Short:        "c",
			Name:         "config",
			Doc:          "[optional] the lint config file",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "d",
			Name:         "detail",
			Doc:          "[optional] print details of lint result",
			DefaultValue: true,
		},
		command.StringFlag{
			Short:        "o",
			Name:         "output",
			Doc:          "[optional] result output file name",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "i",
			Name:         "ignoreBase",
			Doc:          "ignore script which marked baseline",
			DefaultValue: true,
		},
	},
	Run: RunMigrationLint,
}

func RunMigrationLint(ctx *command.Context, input, config string, detail bool, output string, ignoreBase bool) (err error) {
	exitFunc := func(_ int) {
		if err != nil {
			os.Exit(1)
		}
	}
	defer exitFunc(1)

	log.Printf("Erda MySQL Lint the input file or directory: %s", input)
	files := new(walk).walk(input, ".sql").filenames()

	var (
		rulers  = configuration.DefaultRulers()
		lintCfg *configuration.Configuration
	)
	if config != "" {
		lintCfg, err = configuration.FromLocal(config)
		if err != nil {
			return errors.Wrapf(err, "filed to read lint configuration from local")
		}
		rulers, err = lintCfg.Rulers()
		if err != nil {
			return errors.Wrapf(err, "failed to generate lint rulers from lint configuration")
		}
	}

	linter := sqllint.New(rulers...)

	for _, filename := range files {
		var data []byte
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "failed to open file, filename: %s", filename)
		}
		if ignoreBase && isBaseScript(data) {
			continue
		}
		if err = linter.Input(data, filename); err != nil {
			return errors.Wrapf(err, "failed to run Erda MySQL Lint on the SQL script, filename: %s", filename)
		}
	}

	if len(linter.Errors()) == 0 {
		log.Println("Erda MySQL Lint OK")
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

	if detail {
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

func isBaseScript(data []byte) bool {
	return bytes.HasPrefix(data, []byte(baseScriptLabel)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel2)) ||
		bytes.HasPrefix(data, []byte(baseScriptLabel3))
}
