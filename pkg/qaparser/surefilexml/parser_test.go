package surefilexml

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParserSuite(t *testing.T) {
	p := DefaultParser{}
	ts, err := p.Parse("127.0.0.1:9009", "accesskey", "secretkey", "test1", "d1.xml")
	assert.Nil(t, err)
	logrus.Info(ts)
}

func TestParserSuites(t *testing.T) {
	p := DefaultParser{}
	ts, err := p.Parse("127.0.0.1:9009", "accesskey", "secretkey", "test1", "TEST-TestSuite.xml")
	assert.Nil(t, err)

	js, err := json.Marshal(ts)
	assert.Nil(t, err)
	logrus.Info(string(js))
}
