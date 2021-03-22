package testngxml

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/cloudstorage"
	"github.com/erda-project/erda/pkg/qaparser"
	"github.com/erda-project/erda/pkg/qaparser/types"
)

type NgParser struct {
}

func init() {
	logrus.Info("register NGTest Parser to manager")
	(NgParser{}).Register()
}

func (ng NgParser) Register() {
	qaparser.Register(ng, types.NGTest)
}

// parse xml to entity
// 1. get file from cloud storage
// 2. parse
func (NgParser) Parse(endpoint, ak, sk, bucket, objectName string) ([]*apistructs.TestSuite, error) {
	client, err := cloudstorage.New(endpoint, ak, sk)
	if err != nil {
		return nil, errors.Wrap(err, "get cloud storage client")
	}

	byteArray, err := client.DownloadFile(bucket, objectName)
	if err != nil {
		return nil, errors.Wrapf(err, "download filename=%s", objectName)
	}

	var ng *NgTestResult
	if ng, err = Ingest(byteArray); err != nil {
		return nil, err
	}

	return ng.Transfer()
}
