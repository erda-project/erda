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

package es

import (
	"fmt"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/elasticsearch"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

type DescribeESInstancesResponse struct {
	aliyun_resources.ResponsePager
	Instances []elasticsearch.Result
}

func ListByCluster(ctx aliyun_resources.Context,
	page aliyun_resources.PageOption, cluster string) (DescribeESInstancesResponse, error) {
	client, err := elasticsearch.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return DescribeESInstancesResponse{}, err
	}
	request := elasticsearch.CreateListTagResourcesRequest()
	request.Scheme = "https"
	if page.PageSize == nil {
		pagesize := 30
		page.PageSize = &pagesize
	}
	request.Size = requests.NewInteger(*page.PageSize)
	if page.PageNumber == nil {
		pagenum := 1
		page.PageNumber = &pagenum
	}
	request.ResourceType = "INSTANCE"
	request.Page = requests.NewInteger(*page.PageNumber)
	tagKey, tagValue := aliyun_resources.GenClusterTag(cluster)
	request.Tags = fmt.Sprintf("[{\"key\":\"%s\",\"value\":\"%s\"}]", tagKey, tagValue)

	// status:
	//   active:     正常
	//   activating: 生效中
	//   inactive:   冻结
	//   invalid:    失效
	response, err := client.ListTagResources(request)
	if err != nil {
		return DescribeESInstancesResponse{}, err
	}

	ids := []string{}
	for _, tg := range response.TagResources.TagResource {
		ids = append(ids, tg.ResourceId)
	}
	instances := []elasticsearch.Result{}
	for _, id := range ids {
		req := elasticsearch.CreateDescribeInstanceRequest()
		req.Scheme = "https"
		req.InstanceId = id
		resp, err := client.DescribeInstance(req)
		if err != nil {
			return DescribeESInstancesResponse{}, err
		}
		instances = append(instances, resp.Result)
	}
	return DescribeESInstancesResponse{
		ResponsePager: aliyun_resources.ResponsePager{
			TotalCount: response.Headers.XTotalCount,
			PageSize:   *page.PageSize,
			PageNumber: *page.PageNumber,
		},
		Instances: instances,
	}, nil
}

func TagResource(ctx aliyun_resources.Context, cluster string, resourceIDs []string) error {
	if len(resourceIDs) == 0 {
		return nil
	}

	client, err := sdk.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create es client failed, error: %+v", err)
		return err
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "elasticsearch.cn-hangzhou.aliyuncs.com"
	request.Version = "2017-06-13"
	request.PathPattern = "/openapi/tags"
	request.Headers["Content-Type"] = "application/json"

	tagKey, tagValue := aliyun_resources.GenClusterTag(cluster)
	body := fmt.Sprintf(`{
  		"ResourceType":"INSTANCE",
  		"ResourceIds":["%s"],
  		"Tags":[{"key":"%s","value":"%s"}]
		}`, strings.Join(resourceIDs, "\",\""), tagKey, tagValue)
	request.Content = []byte(body)

	_, err = client.ProcessCommonRequest(request)
	if err != nil {
		logrus.Errorf("tag es resource failed, cluster: %s, resource ids: %+v, error: %+v", cluster, resourceIDs, err)
		return err
	}
	return nil
}
