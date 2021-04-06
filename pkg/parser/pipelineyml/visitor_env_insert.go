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

type EnvInsertVisitor struct {
	envs map[string]string
}

func NewEnvInsertVisitor(envs map[string]string) *EnvInsertVisitor {
	return &EnvInsertVisitor{envs: envs}
}

func (v *EnvInsertVisitor) Visit(s *Spec) {
	if v.envs == nil {
		return
	}

	if s.Envs == nil {
		s.Envs = make(map[string]string)
	}

	// insert into struct
	for k, v := range v.envs {
		s.Envs[k] = v
	}

	// encode to new yaml
	_, err := GenerateYml(s)
	if err != nil {
		s.appendError(err)
		return
	}
}
