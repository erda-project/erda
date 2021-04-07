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

package zone

import (
	"sync"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func List(ctx aliyun_resources.Context, regions []string) ([]vpc.Zone, error) {
	allzones := []vpc.Zone{}
	listSch := make(chan listS, 20)
	var wg sync.WaitGroup
	wg.Add(len(regions))
	for _, region := range regions {
		ctx.Region = region
		go func(ctx aliyun_resources.Context) {
			defer func() { wg.Done() }()
			listF(ctx, listSch)
		}(ctx)
	}
	wg.Wait()
	close(listSch)
	for s := range listSch {
		allzones = append(allzones, s.zones...)
	}
	return allzones, nil
}

type listS struct {
	zones []vpc.Zone
}

func listF(ctx aliyun_resources.Context, ch chan listS) {
	zones, err := describeZones(ctx)
	if err != nil {
		logrus.Errorf("failed to describezones: %v", err)
		ch <- listS{}
		return
	}
	ch <- listS{zones: zones}
}

func describeZones(ctx aliyun_resources.Context) ([]vpc.Zone, error) {
	client, err := vpc.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return nil, err
	}
	request := vpc.CreateDescribeZonesRequest()
	request.Scheme = "https"

	response, err := client.DescribeZones(request)
	if err != nil {
		return nil, err
	}
	return response.Zones.Zone, nil
}
