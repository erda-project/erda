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
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ons"
	"github.com/golang-collections/collections/set"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func OverwriteTags(ctx aliyun_resources.Context, items []apistructs.CloudResourceTagItem, tags []string, resourceType aliyun_resources.TagResourceType, instanceID string) error {
	var (
		oldTags     []string
		resourceIDs []string
	)

	tagSet := set.New()

	for i := range items {
		resourceIDs = append(resourceIDs, items[i].ResourceID)
		for k, v := range items[i].OldTags {
			if !tagSet.Has(v) {
				oldTags = append(oldTags, items[i].OldTags[k])
				tagSet.Insert(items[i].OldTags[k])
			}
		}
	}

	// unset old tags
	err := Untag(ctx, resourceIDs, oldTags, resourceType, instanceID)
	if err != nil {
		return err
	}

	// set new tags
	err = TagResource(ctx, resourceIDs, tags, resourceType, instanceID)
	return err
}

func TagResource(ctx aliyun_resources.Context, resourceIDs []string, tags []string, resourceType aliyun_resources.TagResourceType, instanceID string) error {
	if len(resourceIDs) == 0 || len(tags) == 0 {
		logrus.Infof("empty resourceIDs or tags, ignore ons tag")
		return nil
	}

	switch resourceType {
	case aliyun_resources.TagResourceTypeOnsInstanceTag:
	case aliyun_resources.TagResourceTypeOnsTopicTag, aliyun_resources.TagResourceTypeOnsGroupTag:
		if instanceID == "" {
			err := fmt.Errorf("missing instance id for ons topic/group tag, resourceIDs:%+v, tags:%+v", resourceIDs, tags)
			logrus.Errorf(err.Error())
			return err
		}
	default:
		err := fmt.Errorf("tag ons related resource failed, support types:%v invalide resource type: %s",
			[]aliyun_resources.TagResourceType{aliyun_resources.TagResourceTypeVpc, aliyun_resources.TagResourceTypeVsw, aliyun_resources.TagResourceTypeEip}, resourceType)
		logrus.Errorf(err.Error())
		return err
	}

	client, err := ons.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create ons client error:%v", err)
		return err
	}

	request := ons.CreateTagResourcesRequest()
	request.Scheme = "https"

	var tagKV []ons.TagResourcesTag
	for i := range tags {
		tagKV = append(tagKV,
			ons.TagResourcesTag{
				Value: "true",
				Key:   tags[i],
			})
	}
	request.Tag = &tagKV
	request.ResourceId = &resourceIDs
	request.ResourceType = resourceType.String()
	request.InstanceId = instanceID

	_, err = client.TagResources(request)
	if err != nil {
		logrus.Errorf("tag ons resource failed, resourceIDs:%v, tags:%v, error:%v", resourceIDs, tags, err)
		return err
	}
	return nil

}

func Untag(ctx aliyun_resources.Context, resourceIDs []string, keys []string, resourceType aliyun_resources.TagResourceType, instanceID string) error {
	if len(resourceIDs) == 0 || len(keys) == 0 {
		logrus.Infof("ignore ons untag action, empty resourceids or keys")
		return nil
	}

	switch resourceType {
	case aliyun_resources.TagResourceTypeOnsInstanceTag:
	case aliyun_resources.TagResourceTypeOnsTopicTag, aliyun_resources.TagResourceTypeOnsGroupTag:
		if instanceID == "" {
			err := fmt.Errorf("missing instance id for ons topic/group tag, resourceIDs:%+v, tags:%+v", resourceIDs, keys)
			logrus.Errorf(err.Error())
			return err
		}
	default:
		err := fmt.Errorf("untag ons related resource failed, support types:%v invalide resource type: %s",
			[]aliyun_resources.TagResourceType{aliyun_resources.TagResourceTypeVpc, aliyun_resources.TagResourceTypeVsw, aliyun_resources.TagResourceTypeEip}, resourceType)
		logrus.Errorf(err.Error())
		return err
	}

	client, err := ons.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("create ons client error:%v", err)
		return err
	}

	request := ons.CreateUntagResourcesRequest()
	request.Scheme = "https"

	request.ResourceId = &resourceIDs
	request.ResourceType = resourceType.String()
	request.InstanceId = instanceID
	request.TagKey = &keys

	_, err = client.UntagResources(request)
	if err != nil {
		logrus.Errorf("untag resource failed, resourceids:%+v, tags:%+v", resourceIDs, keys)
		return err
	}
	return nil
}
