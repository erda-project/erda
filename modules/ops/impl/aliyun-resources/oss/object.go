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
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"

	aliyun_resources "github.com/erda-project/erda/modules/ops/impl/aliyun-resources"
)

type BucketInfo struct {
	AllObjectTotalSize int64
}

func GetBucketInfo(ctx aliyun_resources.Context, bucketname string, location string) (*BucketInfo, error) {
	endpoint := fmt.Sprintf("http://oss-%s.aliyuncs.com", location)
	accessKeyId := ctx.AccessKeyID
	accessKeySecret := ctx.AccessSecret
	// init
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logrus.Errorf("create oss client error: %+v", err)
		return nil, err
	}
	b, err := client.Bucket(bucketname)
	var totalsize int64

	mark := ""
	for {
		options := []oss.Option{}
		if mark != "" {
			options = append(options, oss.Marker(mark))
		}
		objs, err := b.ListObjects(options...)
		if err != nil {
			return nil, err
		}
		if !objs.IsTruncated {
			break
		}
		mark = objs.NextMarker

		for _, o := range objs.Objects {
			totalsize += o.Size
		}
		logrus.Debugf("bucket: [%s], size: %v G, object num: %v", bucketname, totalsize/1024.0/1024.0, len(objs.Objects))
	}
	return &BucketInfo{AllObjectTotalSize: totalsize}, nil
}

func GetBucketsSize(ctx aliyun_resources.Context, buckets []oss.BucketProperties) (int64, error) {
	var allsize int64
	m := new(sync.Map)
	ch := make(chan struct{}, 20)
	var wg sync.WaitGroup
	wg.Add(len(buckets))
	start := time.Now()
	logrus.Debugf("oss buckets num: %d, buckets: %v", len(buckets), buckets)
	for i, b := range buckets {
		ch <- struct{}{}
		logrus.Debugf("start to get bucket [%v] size for [%s]", i, b.Name)
		go func(index int, b oss.BucketProperties) {
			defer func() {
				<-ch
				wg.Done()
			}()
			info, err := GetBucketInfo(ctx, b.Name, b.Location)
			if err != nil {
				logrus.Errorf("get bucket info failed, bucket:%s, error:%v", b.Name, err)
				return
			}
			m.Store(info.AllObjectTotalSize, struct{}{})
			logrus.Debugf("finish get bucket[%d] [%s] size: %vG", index, b.Name, info.AllObjectTotalSize/1024.0/1024.0)
		}(i, b)
	}
	wg.Wait()
	end := time.Now()
	d := end.Sub(start).Minutes()
	m.Range(func(k interface{}, _ interface{}) bool {
		allsize += k.(int64)
		return true
	})
	logrus.Infof("finish calculate oss bucket size, start:%v, end:%v, spend: [%v min], size:%vG", start, end, d, allsize/1024.0/1024.0)
	return allsize, nil
}
