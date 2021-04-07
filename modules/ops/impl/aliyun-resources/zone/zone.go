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
