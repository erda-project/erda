package region

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func List(ctx aliyun_resources.Context) ([]vpc.Region, error) {
	// 这个接口不需要 regionid 参数, 所以这里写 cn-hangzhou 就行了
	client, err := vpc.NewClientWithAccessKey("cn-hangzhou", ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return nil, err
	}

	request := vpc.CreateDescribeRegionsRequest()
	request.Scheme = "https"

	response, err := client.DescribeRegions(request)
	if err != nil {
		return nil, err
	}
	return response.Regions.Region, nil
}
