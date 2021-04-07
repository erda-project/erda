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

package vpc

import (
	"fmt"
	"time"

	libvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func Create(ctx aliyun_resources.Context, req VPCCreateRequest) (string, error) {
	vpclist, _, err := List(ctx, aliyun_resources.DefaultPageOption, aliyun_resources.ActiveRegionIDs(ctx).VPC, "")
	if err != nil {
		return "", err
	}

	for _, v := range vpclist {
		if v.VpcName == req.Name {
			return "", fmt.Errorf("vpc name:[%s] already exists", req.Name)
		}
	}
	client, err := libvpc.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return "", err
	}
	client.SetReadTimeout(5 * time.Minute)
	client.SetConnectTimeout(3 * time.Minute)

	request := libvpc.CreateCreateVpcRequest()
	request.Scheme = "https"
	request.CidrBlock = req.CidrBlock
	request.VpcName = req.Name
	request.Description = req.Description
	response, err := client.CreateVpc(request)
	if err != nil {
		return "", err
	}
	for i := 0; i < 5; i++ {
		descRequest := libvpc.CreateDescribeVpcsRequest()
		descRequest.Scheme = "https"

		descRequest.VpcId = response.VpcId

		descResponse, err := client.DescribeVpcs(descRequest)
		if err != nil {
			return "", err
		}
		if descResponse.Vpcs.Vpc[0].Status == "Available" {
			break
		}
		time.Sleep(3 * time.Second)
	}
	return response.VpcId, nil
}
