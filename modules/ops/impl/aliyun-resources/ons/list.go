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

package ons

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ons"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

type InstanceOns struct {
	Region string `json:"region"`
	ons.InstanceVO
}

func GetInstanceType(t int) string {
	switch t {
	case 1:
		return "Standard edition instance"
	case 2:
		return "Platinum Edition instance"
	default:
		return ""
	}
}

func GetInstanceStatus(status int) string {
	switch status {
	case 0:
		return "Platinum Edition instance deploying"
	case 2:
		return "Standard edition instance expired"
	case 5:
		return "Running"
	case 7:
		return "Upgrading"
	default:
		return ""
	}
}

func GetMsgType(msgType int) string {
	switch msgType {
	case 0:
		return "Normal message"
	case 1:
		return "Partitionally ordered message"
	case 2:
		return "Globally ordered message"
	case 4:
		return "Transactional Message"
	case 5:
		return "Scheduled/delayed message"
	default:
		return ""
	}
}

func List(ctx aliyun_resources.Context, page aliyun_resources.PageOption,
	regions []string,
	_cluster string) ([]InstanceOns, error) {
	var resultList []InstanceOns
	for _, region := range regions {
		ctx.Region = region
		items, err := DescribeResource(ctx)
		if err != nil {
			logrus.Errorf("describe ons failed, error:%v", err)
			return nil, err
		}
		resultList = append(resultList, items...)
	}
	return resultList, nil
}

func DescribeResource(ctx aliyun_resources.Context) ([]InstanceOns, error) {
	client, err := ons.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return nil, err
	}

	request := ons.CreateOnsInstanceInServiceListRequest()
	request.Scheme = "https"

	response, err := client.OnsInstanceInServiceList(request)
	if err != nil {
		return nil, err
	}
	var result []InstanceOns
	for i := range response.Data.InstanceVO {
		result = append(result, InstanceOns{
			Region:     ctx.Region,
			InstanceVO: response.Data.InstanceVO[i],
		})
	}
	return result, nil
}

func DescribeTopic(ctx aliyun_resources.Context, req apistructs.CloudResourceOnsTopicInfoRequest) ([]ons.PublishInfoDo, error) {
	client, err := ons.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return nil, err
	}

	request := ons.CreateOnsTopicListRequest()
	request.Scheme = "https"

	request.InstanceId = req.InstanceID
	if req.TopicName != "" {
		request.Topic = req.TopicName
	}

	response, err := client.OnsTopicList(request)
	if err != nil {
		return nil, err
	}
	return response.Data.PublishInfoDo, nil
}

func DescribeGroup(ctx aliyun_resources.Context, req apistructs.CloudResourceOnsGroupInfoRequest) ([]ons.SubscribeInfoDo, error) {
	client, err := ons.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create client failed, error:%v", err)
		return nil, err
	}

	request := ons.CreateOnsGroupListRequest()
	request.Scheme = "https"

	request.InstanceId = req.InstanceID
	if req.GroupID != "" {
		request.GroupId = req.GroupID
	}
	if req.GroupType != "" {
		request.GroupType = req.GroupType
	}

	response, err := client.OnsGroupList(request)
	if err != nil {
		logrus.Errorf("describe ons group failed, error:%v", err)
		return nil, err
	}
	return response.Data.SubscribeInfoDo, nil
}

func Classify(ins []InstanceOns) (runningCount, gonnaExpiredCount, expiredCount, stoppedCount,
	postpaidCount, prepaidCount int, err error) {
	now := time.Now().UTC()
	for _, v := range ins {
		if v.InstanceType == 1 {
			postpaidCount += 1
		} else {
			prepaidCount += 1
		}

		if v.InstanceType == 1 {
			runningCount += 1
			continue
		}

		// expire time only apply to post paid instance
		var t time.Time
		t = time.Unix(v.ReleaseTime, 0).UTC()
		if t.Before(now) {
			expiredCount += 1
		} else if t.Before(now.Add(24 * 10 * time.Hour)) {
			gonnaExpiredCount += 1
		} else {
			runningCount += 1
		}
	}
	return
}
