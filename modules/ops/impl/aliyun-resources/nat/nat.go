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

package nat

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func DescribeResource(ctx aliyun_resources.Context,
	page aliyun_resources.PageOption, natGatewayId string) (*vpc.DescribeNatGatewaysResponse, error) {

	client, err := vpc.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create vpc client error: %+v", err)
		return nil, err
	}
	request := vpc.CreateDescribeNatGatewaysRequest()
	request.Scheme = "https"
	request.RegionId = ctx.Region
	request.NatGatewayId = natGatewayId

	response, err := client.DescribeNatGateways(request)
	if err != nil {
		logrus.Errorf("describe nat failed, error: %+v", err)
		return nil, err
	}
	return response, nil
}

func DescribeSnatEntry(ctx aliyun_resources.Context,
	page aliyun_resources.PageOption, snatTableId string) (*vpc.DescribeSnatTableEntriesResponse, error) {
	client, err := vpc.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create vpc client error: %+v", err)
		return nil, err
	}

	request := vpc.CreateDescribeSnatTableEntriesRequest()
	request.Scheme = "https"

	request.SnatTableId = snatTableId
	request.PageSize = requests.NewInteger(*page.PageSize)

	response, err := client.DescribeSnatTableEntries(request)
	if err != nil {
		logrus.Errorf("describe snat table failed, error: %+v", err)
		return nil, err
	}
	return response, nil
}

func IsVswitchBoundSnat(ctx aliyun_resources.Context, snatTableId string, vswid string) (bool, error) {
	if vswid == "" {
		return false, nil
	}
	page := aliyun_resources.DefaultPageOption
	rsp, err := DescribeSnatEntry(ctx, page, snatTableId)
	if err != nil {
		logrus.Errorf("describe snat entry failed, err:%v", err)
		return false, err
	}
	if rsp == nil {
		err := fmt.Errorf("describe snat entry failed, empty response")
		logrus.Errorf(err.Error())
		return false, err
	}
	for _, v := range rsp.SnatTableEntries.SnatTableEntry {
		if v.SourceVSwitchId == vswid {
			return true, nil
		}
	}
	return false, nil
}
