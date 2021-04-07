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
