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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type VersionVisitor struct{}

func NewVersionVisitor() *VersionVisitor {
	return &VersionVisitor{}
}

func (v *VersionVisitor) Visit(s *Spec) {
	switch s.Version {
	case "":
		s.appendError(errors.New("no version"))
	case Version1dot1:
		return
	default:
		s.appendError(errors.Errorf("invalid version: %s, only support 1.1", s.Version))
	}
}

func GetVersion(data []byte) (string, error) {
	type version struct {
		Version string `yaml:"version"`
	}
	var ver version
	if err := yaml.Unmarshal(data, &ver); err != nil {
		return "", err
	}
	return ver.Version, nil
}
