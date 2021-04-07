package redis

import (
	kvstore "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"github.com/golang-collections/collections/set"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

func OverwriteTags(ctx aliyun_resources.Context, items []apistructs.CloudResourceTagItem, tags []string) error {
	var (
		oldTags     []string
		instanceIDs []string
	)

	tagSet := set.New()

	for i := range items {
		instanceIDs = append(instanceIDs, items[i].ResourceID)
		for k, v := range items[i].OldTags {
			if !tagSet.Has(v) {
				oldTags = append(oldTags, items[i].OldTags[k])
				tagSet.Insert(items[i].OldTags[k])
			}
		}
	}

	// unset old tags
	err := UnTag(ctx, instanceIDs, oldTags)
	if err != nil {
		return err
	}

	// set new tags
	err = TagResource(ctx, instanceIDs, tags)
	return err
}

func TagResource(ctx aliyun_resources.Context, instanceIds []string, tags []string) error {
	if len(instanceIds) == 0 || len(tags) == 0 {
		return nil
	}

	// get resource detail info with tags
	client, err := kvstore.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("redis, create client failed: %+v", err)
		return err
	}

	request := kvstore.CreateTagResourcesRequest()
	request.Scheme = "https"
	request.ResourceType = "INSTANCE"
	request.ResourceId = &instanceIds

	var tagKV []kvstore.TagResourcesTag
	for i := range tags {
		tagKV = append(tagKV, kvstore.TagResourcesTag{
			Value: "true",
			Key:   tags[i],
		})
	}
	request.Tag = &tagKV

	_, err = client.TagResources(request)
	if err != nil {
		logrus.Errorf("redis, tag resource failed, error:%v", err)
		return err
	}
	return nil
}

func UnTag(ctx aliyun_resources.Context, instanceIds []string, tags []string) error {
	if len(instanceIds) == 0 || len(tags) == 0 {
		return nil
	}

	client, err := kvstore.NewClientWithAccessKey(ctx.Region, ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		logrus.Errorf("redis, create client failed: %+v", err)
		return err
	}

	request := kvstore.CreateUntagResourcesRequest()
	request.Scheme = "https"

	request.ResourceType = "instance"
	request.ResourceId = &instanceIds
	if len(tags) > 0 {
		request.TagKey = &tags
	}

	_, err = client.UntagResources(request)
	if err != nil {
		logrus.Errorf("redis, failed to untag resource, instances ids:%v, tags:%v, error:%v", instanceIds, tags, err)
		return err
	}
	return nil
}
