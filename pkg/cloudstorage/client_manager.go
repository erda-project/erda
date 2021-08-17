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

package cloudstorage

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/cloudstorage/minioclient"
	"github.com/erda-project/erda/pkg/cloudstorage/ossclient"
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
