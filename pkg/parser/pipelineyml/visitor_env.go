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

package pipelineyml

import (
	"regexp"

	"github.com/pkg/errors"
)

var envKeyRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

type EnvVisitor struct {
	envs map[string]string
}

func NewEnvVisitor(envs map[string]string) *EnvVisitor {
	return &EnvVisitor{envs: envs}
}

func (v *EnvVisitor) Visit(s *Spec) {
	if s.Envs == nil {
		s.Envs = make(map[string]string, 0)
	}

	for k, v := range v.envs {
		s.Envs[k] = v
	}

	checkErr := CheckEnvs(s.Envs)
	for _, err := range checkErr {
		s.appendError(err)
	}
}

func CheckEnvs(envs map[string]string) []error {
	var errs []error
	for k := range envs {
		if !envKeyRegexp.MatchString(k) {
			errs = append(errs, errors.Errorf("invalid env key: %s (must match: %s)", k, envKeyRegexp.String()))
		}
	}
	return errs
}
