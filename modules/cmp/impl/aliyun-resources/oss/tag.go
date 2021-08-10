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

package oss

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/golang-collections/collections/set"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	aliyun_resources "github.com/erda-project/erda/modules/cmp/impl/aliyun-resources"
)

func OverwriteTags(ctx aliyun_resources.Context, items []apistructs.CloudResourceTagItem, tags []string) error {
	var (
		oldTags     []string
		instanceIDs []string
	)

	tagSet := set.New()

	// merge old tags, unset non exist tags is ok
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

func TagResource(ctx aliyun_resources.Context, buckets []string, tags []string) error {
	if len(buckets) == 0 || len(tags) == 0 {
		logrus.Infof("ignore oss tag request, empty buckets or tags")
		return nil
	}

	// create client
	endpoint := fmt.Sprintf("http://oss-%s.aliyuncs.com", ctx.Region)
	accessKeyId := ctx.AccessKeyID
	accessKeySecret := ctx.AccessSecret
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logrus.Errorf("create oss client failed, error:%v", err)
		return err
	}

	tagKV := oss.Tagging{}
	for _, tagK := range tags {
		tagKV.Tags = append(tagKV.Tags, oss.Tag{Key: tagK, Value: "true"})
	}

	// set bucket tag
	m := new(sync.Map)
	var wg sync.WaitGroup
	for _, b := range buckets {
		wg.Add(1)
		go func(bucketName string) {
			defer func() {
				wg.Done()
			}()
			logrus.Infof("start to tag bucket:%s", bucketName)
			err := client.SetBucketTagging(bucketName, tagKV)
			if err != nil {
				logrus.Errorf("set tag for oss bucket failed, bucket name:%s, error:%v", bucketName, err)
				m.Store(bucketName, struct{}{})
				return
			}
		}(b)
	}
	wg.Wait()
	var failedBuckets []string
	m.Range(func(k interface{}, _ interface{}) bool {
		failedBuckets = append(failedBuckets, k.(string))
		return true
	})
	if len(failedBuckets) > 0 {
		err := fmt.Errorf("batch set tag for oss bucket failed, failed buckets:%v", failedBuckets)
		logrus.Errorf(err.Error())
		return err
	}
	return nil
}

func GetResourceTags(ctx aliyun_resources.Context, buckets []string) (*map[string]oss.Tagging, error) {
	// create client
	endpoint := fmt.Sprintf("http://oss-%s.aliyuncs.com", ctx.Region)
	accessKeyId := ctx.AccessKeyID
	accessKeySecret := ctx.AccessSecret
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logrus.Errorf("create oss client failed, error:%v", err)
		return nil, err
	}
	// get bucket tags
	mBucketsInfo := new(sync.Map)
	mFailedBuckets := new(sync.Map)
	var wg sync.WaitGroup
	for _, b := range buckets {
		wg.Add(1)
		go func(bucketName string) {
			defer func() {
				wg.Done()
			}()
			result, err := client.GetBucketTagging(bucketName)
			if err != nil {
				logrus.Errorf("get tag for oss bucket failed, bucket name:%s, error:%v", bucketName, err)
				mFailedBuckets.Store(bucketName, struct{}{})
				return
			}
			if len(result.Tags) > 0 {
				logrus.Infof("bucket:%s, with tag:%+v", bucketName, result)
				mBucketsInfo.Store(bucketName, result)
			}
		}(b)
	}
	wg.Wait()
	var failedBuckets []string
	mFailedBuckets.Range(func(k interface{}, _ interface{}) bool {
		failedBuckets = append(failedBuckets, k.(string))
		return true
	})
	if len(failedBuckets) > 0 {
		err := fmt.Errorf("batch set tag for oss bucket failed, failed buckets:%v", failedBuckets)
		logrus.Errorf(err.Error())
		return nil, err
	}
	bucketsInfo := make(map[string]oss.Tagging)
	mBucketsInfo.Range(func(key, value interface{}) bool {
		bucketsInfo[key.(string)] = oss.Tagging(value.(oss.GetBucketTaggingResult))
		return true
	})

	return &bucketsInfo, nil
}

func UnTag(ctx aliyun_resources.Context, buckets []string, tags []string) error {
	// pre check
	if len(buckets) == 0 || len(tags) == 0 {
		return nil
	}

	// create client
	endpoint := fmt.Sprintf("http://oss-%s.aliyuncs.com", ctx.Region)
	accessKeyId := ctx.AccessKeyID
	accessKeySecret := ctx.AccessSecret
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logrus.Errorf("create oss client failed, error:%v", err)
		return err
	}

	var opts []oss.Option
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		opts = append(opts, oss.TagKey(tag))
	}

	// set bucket tag
	m := new(sync.Map)
	var wg sync.WaitGroup
	for _, b := range buckets {
		wg.Add(1)
		go func(bucketName string) {
			defer func() {
				wg.Done()
			}()
			err := client.DeleteBucketTagging(bucketName, opts...)
			if err != nil {
				logrus.Errorf("untag oss bucket failed, bucket name:%s, error:%v", bucketName, err)
				m.Store(bucketName, struct{}{})
				return
			}
		}(b)
	}
	wg.Wait()
	var failedBuckets []string
	m.Range(func(k interface{}, _ interface{}) bool {
		failedBuckets = append(failedBuckets, k.(string))
		return true
	})
	if len(failedBuckets) > 0 {
		err := fmt.Errorf("batch untag oss bucket failed, failed buckets:%v", failedBuckets)
		logrus.Errorf(err.Error())
		return err
	}
	return nil

}
