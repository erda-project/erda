package surefilexml

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/cloudstorage"
	"github.com/erda-project/erda/pkg/qaparser"
	"github.com/erda-project/erda/pkg/qaparser/types"
)

type DefaultParser struct {
}

// parse xml to entity
// 1. get file from cloud storage
// 2. parse
func init() {
	logrus.Info("register Default Parser to manager")
	(DefaultParser{}).Register()
}

func (d DefaultParser) Register() {
	qaparser.Register(d, types.Default, types.JUnit)
}

func (DefaultParser) Parse(endpoint, ak, sk, bucket, objectName string) ([]*apistructs.TestSuite, error) {
	client, err := cloudstorage.New(endpoint, ak, sk)
	if err != nil {
		return nil, errors.Wrap(err, "get cloud storage client")
	}

	byteArray, err := client.DownloadFile(bucket, objectName)
	if err != nil {
		return nil, errors.Wrapf(err, "download filename=%s", objectName)
	}

	var suites []*apistructs.TestSuite
	if suites, err = Ingest(byteArray); err != nil {
		return nil, err
	}

	return suites, nil
}
