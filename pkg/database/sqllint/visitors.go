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

package sqllint

import (
	"github.com/pingcap/parser"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type Linter struct {
	stop    bool
	layer   int
	errs    map[string][]error
	reports map[string]map[string][]string
	linters []rules.Ruler
}

func New(rules ...rules.Ruler) *Linter {
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

	s := script.New(scriptName, scriptData)
	r.reports[scriptName] = make(map[string][]string, 0)

	var errs []error
	for _, node := range nodes {
		for _, f := range r.linters {
			linter := f(s)
			_, _ = node.Accept(linter)
			if err := linter.Error(); err != nil {
				errs = append(errs, err)
				lintError, ok := err.(linterror.LintError)
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
