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
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type scriptChecksumLinter struct {
	baseLinter
	meta []scriptChecksumLinterMetaItem
}

func (hub) ScriptChecksumLinter(s script.Script, c sqllint.Config) (sqllint.Rule, error) {
	var l = scriptChecksumLinter{
		baseLinter: newBaseLinter(s),
		meta:       nil,
	}
	if err := yaml.Unmarshal(c.Meta, &l.meta); err != nil {
		return nil, errors.Wrap(err, "failed to parse ScriptNameLinter.meta")
	}
	for _, item := range l.meta {
		if item.ScriptName == "" {
			return nil, errors.New("invali checksum item in ScriptChecksumLinter.meta: empty scriptName")
		}
		if item.Checksum == "" {
			return nil, errors.New("invalid checksum item in ScriptChecksumLinter.meta: empty checksum")
		}
	}
	return &l, nil
}

func (l *scriptChecksumLinter) LintOnScript() {
	scriptExt := path.Ext(l.s.Name())
	scriptName := strings.TrimSuffix(l.s.Name(), scriptExt)
	for _, item := range l.meta {
		itemExt := path.Ext(item.ScriptName)
		itemName := strings.TrimSuffix(item.ScriptName, itemExt)
		if (itemExt == "" && scriptName == itemName) || (itemExt != "" && l.s.Name() == item.ScriptName) {
			if item.Checksum != l.s.Checksum() {
				l.err = linterror.New(l.s, l.s.Name(),
					fmt.Sprintf("invalid script checksum, it is expected to be: %s, got: %s", item.Checksum, l.s.Checksum()),
					func(_ []byte) bool { return true })
				return
			}
		}
	}
}

type scriptChecksumLinterMetaItem struct {
	ScriptName string `json:"scriptName" yaml:"scriptName"`
	Checksum   string `json:"checksum" yaml:"checksum"`
}
