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

package sqllint

import (
	"github.com/pingcap/parser"
	"gopkg.in/yaml.v3"
)

type Linter struct {
	stop    bool
	layer   int
	errs    map[string][]error
	reports map[string]map[string][]string
	linters []NewRule
}

func New(rules ...NewRule) *Linter {
	r := &Linter{
		stop:    false,
		layer:   0,
		errs:    make(map[string][]error, 0),
		reports: make(map[string]map[string][]string, 0),
		linters: nil,
	}
	for _, l := range rules {
		r.linters = append(r.linters, l)
	}
	return r
}

func (r *Linter) Input(scriptData []byte, scriptName string) error {
	p := parser.New()
	nodes, warns, err := p.Parse(string(scriptData), "", "")
	if err != nil {
		return err
	}

	script := NewScript(scriptName, scriptData)
	r.reports[scriptName] = make(map[string][]string, 0)

	var errs []error
	for _, node := range nodes {
		for _, f := range r.linters {
			linter := f(script)
			_, _ = node.Accept(linter)
			if err := linter.Error(); err != nil {
				errs = append(errs, err)
				lintError, ok := err.(LintError)
				if !ok {
					continue
				}
				stmtName := lintError.StmtName()
				if stmtName == "" {
					continue
				}
				r.reports[scriptName][stmtName] = append(r.reports[scriptName][stmtName], lintError.Lint)
			}
		}
	}

	if len(warns) > 0 {
		r.errs[scriptName+" [warns]"] = warns
	}
	if len(errs) > 0 {
		r.errs[scriptName+" [lints]"] = errs
	}

	return nil
}

func (r *Linter) Errors() map[string][]error {
	return r.errs
}

func (r *Linter) Report() string {
	data, err := yaml.Marshal(r.reports)
	if err != nil {
		return ""
	}
	return string(data)
}
