package cloudstorage

import (
	"fmt"

	"github.com/erda-project/erda/pkg/cloudstorage/minioclient"
	"github.com/erda-project/erda/pkg/cloudstorage/ossclient"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client interface {
	UploadFile(bucketName, objectName, file string) (string, error)
	DownloadFile(bucketName, objectName string) ([]byte, error)
	GetFileUrl(bucketName, objectName string) (string, error)
	HealthCheck() error
}

func New(endpoint, accessKey, secretKey string) (Client, error) {
	var client Client
	var err error

	// return oss client if connected
	client, err = ossclient.New(endpoint, accessKey, secretKey)
	if err == nil {
		return client, nil
	}

	// return minio client if connected
	client, err = minioclient.New(endpoint, accessKey, secretKey)
	if err == nil {
		return client, nil
	}

	errMsg := fmt.Sprintf("get cloud storage failed. Either oss or minio client found by endpoint=%s", endpoint)
	logrus.Errorf(errMsg)
	return nil, errors.Errorf(errMsg)
}
