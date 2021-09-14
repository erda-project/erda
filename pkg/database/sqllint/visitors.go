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
	"fmt"
	"io"
	"os"

	"github.com/pingcap/parser"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/database/schema"
	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type Linter struct {
	stop    bool
	layer   int
	errs    map[string]*Errors
	reports map[string]map[string][]string
	linters []rules.Ruler
	schema  *schema.Schema
}

func New(rules ...rules.Ruler) *Linter {
	r := &Linter{
		stop:    false,
		layer:   0,
		errs:    make(map[string]*Errors),
		reports: make(map[string]map[string][]string, 0),
		linters: nil,
		schema:  schema.New(),
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
	// update schema state
	for _, node := range nodes {
		node.Accept(r.schema)
	}
	s := script.New(scriptName, scriptData)
	r.reports[scriptName] = make(map[string][]string, 0)

	var errs []error
	for _, node := range nodes {
		for _, f := range r.linters {
			linter := f(s)
			// if the linter lints with state, set its schema
			if stateLinter, ok := linter.(rules.StatefulRule); ok {
				stateLinter.SetSchema(r.schema)
			}
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

	if len(warns) > 0 || len(errs) > 0 {
		if r.errs[scriptName] == nil {
			r.errs[scriptName] = new(Errors)
		}
		r.errs[scriptName].Warns = append(r.errs[scriptName].Warns, warns...)
		r.errs[scriptName].Lints = append(r.errs[scriptName].Lints, errs...)
	}

	return nil
}

func (r *Linter) Errors() map[string]*Errors {
	return r.errs
}

func (r *Linter) GetError(scriptName string) *Errors {
	return r.errs[scriptName]
}

func (r *Linter) HasError(scriptNames ...string) bool {
	var names = make(map[string]bool, len(scriptNames))
	for _, name := range scriptNames {
		names[name] = true
	}
	for k, v := range r.errs {
		if len(v.Lints) > 0 && (len(names) == 0 || names[k]) {
			return true
		}
	}
	return false
}

func (r *Linter) Report() string {
	data, err := yaml.Marshal(r.reports)
	if err != nil {
		return ""
	}
	return string(data)
}

// FprintErrors writes errors details to the w io.Writer
func (r *Linter) FprintErrors(w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	_, _ = fmt.Fprintln(w, r.Report())
	errors := r.Errors()
	for filename, es := range errors {
		_, _ = fmt.Fprintln(w, filename)
		for i := range es.Lints {
			_, _ = fmt.Fprintf(w, "ERRORS[%v]\n%v", i, es.Lints[i])
		}
		for i := range es.Warns {
			_, _ = fmt.Fprintf(w, "WARNS[%v]\n%v", i, es.Warns[i])
		}
	}
}

type Errors struct {
	Warns []error
	Lints []error
}
