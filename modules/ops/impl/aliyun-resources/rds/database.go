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

package rds

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

// describe database info
func DescribeDatabases(ctx aliyun_resources.Context, req apistructs.CloudResourceMysqlListDatabaseRequest) (*rds.DescribeDatabasesResponse, error) {
	client, err := rds.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create rds client error: %v", err)
		return nil, err
	}

	request := rds.CreateDescribeDatabasesRequest()
	request.Scheme = "https"
	request.DBInstanceId = req.InstanceID

	response, err := client.DescribeDatabases(request)
	if err != nil {
		logrus.Errorf("describe mysql database failed, error:%v", err)
		return nil, err
	}
	return response, nil
}
