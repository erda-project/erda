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

package nas

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/nas"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

type DescribeFileSystemResponse struct {
	aliyun_resources.ResponsePager
	FileSystems []nas.DescribeFileSystemsFileSystem1
}

func ListByCluster(ctx aliyun_resources.Context,
	page aliyun_resources.PageOption, cluster string) (DescribeFileSystemResponse, error) {
	client, err := nas.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return DescribeFileSystemResponse{}, err
	}
	request := nas.CreateDescribeTagsRequest()
	request.Scheme = "https"
	if page.PageSize != nil {
		request.PageSize = requests.NewInteger(*page.PageSize)
	}
	if page.PageNumber != nil {
		request.PageNumber = requests.NewInteger(*page.PageNumber)
	}
	tagKey, tagValue := aliyun_resources.GenClusterTag(cluster)
	request.Tag = &[]nas.DescribeTagsTag{
		{
			Key:   tagKey,
			Value: tagValue,
		},
	}
	response, err := client.DescribeTags(request)
	if err != nil {
		return DescribeFileSystemResponse{}, err
	}

	if len(response.Tags.Tag) == 0 {
		return DescribeFileSystemResponse{
			ResponsePager: aliyun_resources.ResponsePager{
				TotalCount: response.TotalCount,
				PageSize:   response.PageSize,
				PageNumber: response.PageNumber,
			},
			FileSystems: nil,
		}, nil
	}
	ids := response.Tags.Tag[0].FileSystemIds.FileSystemId

	fslist := []nas.DescribeFileSystemsFileSystem1{}

	for _, id := range ids {
		request := nas.CreateDescribeFileSystemsRequest()
		request.Scheme = "https"
		request.FileSystemId = id

		// status
		//  Pending: 等待
		//  Running: 运行
		//  Stopped: 停止
		resp, err := client.DescribeFileSystems(request)
		if err != nil {
			return DescribeFileSystemResponse{}, err
		}
		fs := resp.FileSystems.FileSystem
		if len(fs) == 0 {
			return DescribeFileSystemResponse{},
				fmt.Errorf("failed to Describefilesystem, fsid: %s, regionid: %s, ak: %s",
					id, ctx.Region, ctx.AccessKeyID)
		}
		fslist = append(fslist, fs[0])
	}
	return DescribeFileSystemResponse{
		ResponsePager: aliyun_resources.ResponsePager{
			TotalCount: response.TotalCount,
			PageSize:   response.PageSize,
			PageNumber: response.PageNumber,
		},
		FileSystems: fslist,
	}, nil
}

func TagResource(ctx aliyun_resources.Context, cluster string, resourceIDs []string) error {
	if len(resourceIDs) == 0 {
		return nil
	}

	for _, id := range resourceIDs {
		_ = TagOneResource(ctx, cluster, id)
	}

	return nil
}

func TagOneResource(ctx aliyun_resources.Context, cluster string, resourceID string) error {
	client, err := nas.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create rds client error: %+v", err)
	}

	request := nas.CreateAddTagsRequest()
	request.Scheme = "https"
	request.RegionId = ctx.Region
	tagKey, tagValue := aliyun_resources.GenClusterTag(cluster)
	request.Tag = &[]nas.AddTagsTag{{Key: tagKey, Value: tagValue}}
	request.FileSystemId = resourceID

	_, err = client.AddTags(request)
	if err != nil {
		logrus.Errorf("tag nas resource failed, cluster: %s, resource ids: %+v, error: %+v", cluster, resourceID, err)
		return err
	}
	return nil
}
