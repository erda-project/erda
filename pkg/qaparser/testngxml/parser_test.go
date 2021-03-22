package testngxml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	filename := "testng-results.xml"
	r, err := (NgParser{}).Parse("127.0.0.1:9009", "accesskey", "secretkey", "test1", filename)
	assert.Nil(t, err)

	js, err := json.Marshal(r)
	assert.Nil(t, err)
	logrus.Info(string(js))
}

func TestIngest(t *testing.T) {
	bs, err := ioutil.ReadFile("../testdata/testng-results.xml")
	assert.Nil(t, err)

	ng, err := Ingest(bs)
	assert.Nil(t, err)

	js, _ := json.Marshal(ng)
	logrus.Info(string(js))

	fmt.Println(string(js))

}
