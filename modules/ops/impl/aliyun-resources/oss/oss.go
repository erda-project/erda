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
	"sort"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

type ossBuckets []oss.BucketProperties

func (o ossBuckets) Len() int {
	return len(o)
}

func (o ossBuckets) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o ossBuckets) Less(i, j int) bool {
	return o[i].CreationDate.After(o[j].CreationDate)
}

func List(ctx aliyun_resources.Context, page aliyun_resources.PageOption,
	regions []string,
	_cluster string,
	tags []string,
	prefix string) ([]oss.BucketProperties, error) {
	bucketList := []oss.BucketProperties{}
	// oss, it will return buckets in all regions when offer one region.
	regions = []string{"cn-hangzhou"}
	for _, region := range regions {
		ctx.Region = region
		buckets, err := DescribeResource(ctx, page, _cluster, tags, prefix)
		if err != nil {
			logrus.Errorf("describe resource failed, %+v", err)
			return nil, err
		}
		bucketList = append(bucketList, buckets...)
	}
	sort.Sort(ossBuckets(bucketList))
	return bucketList, nil
}

func DescribeResource(ctx aliyun_resources.Context, page aliyun_resources.PageOption,
	_cluster string, tags []string, prefix string) ([]oss.BucketProperties, error) {
	endpoint := fmt.Sprintf("http://oss-%s.aliyuncs.com", ctx.Region)
	accessKeyId := ctx.AccessKeyID
	accessKeySecret := ctx.AccessSecret
	// init
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logrus.Errorf("create oss client error: %+v", err)
		return nil, err
	}
	// request
	options := []oss.Option{}
	// set max bucket
	maxBucket := 100
	maxkeyOption := oss.MaxKeys(maxBucket)
	options = append(options, maxkeyOption)
	// set prefix
	if prefix != "" {
		prefixOption := oss.Prefix(prefix)
		options = append(options, prefixOption)
	}
	// set filter tags
	// the relationship between multiple tags is and in oss.
	for _, v := range tags {
		options = append(options, oss.TagKey(v))
	}
	rsp, err := client.ListBuckets(options...)
	if err != nil {
		logrus.Errorf("list bucket error:%v", err)
	}
	for i, v := range rsp.Buckets {
		rsp.Buckets[i].Location = strings.TrimPrefix(v.Location, "oss-")
	}
	return rsp.Buckets, nil
}
