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
