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

package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type scriptNameLinter struct {
	baseLinter
	meta scriptNameLinterMeta
	regs []*regexp.Regexp
}

func (hub) ScriptNameLinter(s script.Script, c sqllint.Config) (sqllint.Rule, error) {
	var l = scriptNameLinter{
		baseLinter: newBaseLinter(s),
		meta:       scriptNameLinterMeta{},
	}
	if err := yaml.Unmarshal(c.Meta, &l.meta); err != nil {
		return nil, errors.Wrap(err, "failed to parse ScriptNameLinter.meta")
	}
	if len(l.meta.Patterns) == 0 {
		return nil, errors.New("no script name pattern in ScriptNameLinter.meta")
	}
	for _, pat := range l.meta.Patterns {
		reg, err := regexp.Compile(pat)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid pattern: %s", pat)
		}
		l.regs = append(l.regs, reg)
	}
	return &l, nil
}

func (l *scriptNameLinter) LintOnScript() {
	for _, pat := range l.meta.Patterns {
		if ok, err := regexp.MatchString(pat, l.s.Name()); err == nil && ok {
			return
		}
	}
	l.err = linterror.New(l.s, l.s.Name(),
		fmt.Sprintf("invalid script name, it should match one of these patterns:\n%s", strings.Join(l.meta.Patterns, "\n")),
		func(_ []byte) bool { return true })
}

type scriptNameLinterMeta struct {
	Patterns []string `json:"patterns" yaml:"patterns"`
}
