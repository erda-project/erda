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

package qaparser

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/qaparser/types"
)

var m Manager

type Parser interface {
	Parse(endpoint, ak, sk, bucket, objectName string) ([]*apistructs.TestSuite, error)
	Register()
}

type Manager struct {
	parsers map[types.TestParserType]Parser
}

func GetManager() *Manager {
	return &m
}

func init() {
	m = Manager{
		parsers: map[types.TestParserType]Parser{},
	}

	logrus.Info(">>> init parser manager finished <<<")
}

func (m *Manager) GetParser(t types.TestParserType) Parser {

	if _, ok := m.parsers[t]; !ok {
		logrus.Errorf(">>> not found test type=%s <<<", t)
	}
	return m.parsers[t]
}

func Register(p Parser, types ...types.TestParserType) error {
	for _, t := range types {
		if _, ok := m.parsers[t]; ok {
			return errors.Errorf("duplicated type=%s", t.TPValue())
		}
		logrus.Infof(">>> register type=%s to parser manager success <<<", t)
		m.parsers[t] = p
	}

	return nil
}
