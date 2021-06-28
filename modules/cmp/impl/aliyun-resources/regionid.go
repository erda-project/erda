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

package aliyun_resources

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	libvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/golang-collections/collections/set"
	"github.com/sirupsen/logrus"
)

type CachedRegionIDs struct {
	ECS                 []string
	VPC                 []string
	LastUpdateTimestamp int64
}

var globalRegionIDs CachedRegionIDs
var vpcRegionIDs, ecsRegionIDs *set.Set

func allRegionIDs(ctx Context) ([]string, error) {
	client, err := libvpc.NewClientWithAccessKey("cn-hangzhou", ctx.AccessKeyID, ctx.AccessSecret)
	if err != nil {
		return nil, err
	}
	request := libvpc.CreateDescribeRegionsRequest()
	request.Scheme = "https"
	response, err := client.DescribeRegions(request)
	if err != nil {
		return nil, err
	}
	regionIDs := []string{}
	for _, r := range response.Regions.Region {
		// ignore regions, always connect timeout: "cn-wulanchabu"
		if r.RegionId == "cn-wulanchabu" {
			continue
		}
		regionIDs = append(regionIDs, r.RegionId)
	}
	return regionIDs, nil
}

func ActiveRegionIDs(ctx Context) CachedRegionIDs {
	now := time.Now()
	// try to get org cache etcd regions
	if regions := GetCachedRegions(ctx); regions != nil {
		if now.Unix()-regions.LastUpdateTimestamp >= 10*60 {
			// update regions async
			go func() {
				newRegions := refresh(ctx)
				PutCachedRegions(ctx, newRegions)
				UpdateGlobalCachedRegions(newRegions)
			}()
		}
		logrus.Infof("get regions from etcd cache:%+v", *regions)
		return *regions
	}
	// try to get global cached regions
	if len(globalRegionIDs.VPC) > 0 {
		logrus.Infof("get regions from from global cache:%+v", globalRegionIDs)
		return globalRegionIDs
	}
	newRegions := refresh(ctx)
	PutCachedRegions(ctx, newRegions)
	UpdateGlobalCachedRegions(newRegions)
	if newRegions == nil || len(newRegions.VPC) == 0 {
		logrus.Infof("get regions from from global cache:%+v", globalRegionIDs)
		return globalRegionIDs
	}
	logrus.Infof("get regions from refresh:%+v", *newRegions)
	return *newRegions
}

func refresh(ctx Context) *CachedRegionIDs {
	allRegions, err := allRegionIDs(ctx)
	if err != nil {
		logrus.Errorf("failed to get allRegionIDs: %v", err)
		return nil
	}
	filterRegionIDs := filterRegionIDs(ctx, CachedRegionIDs{
		ECS: allRegions,
		VPC: allRegions,
	})
	return &filterRegionIDs
}

func filterRegionIDs(ctx Context, regionids CachedRegionIDs) CachedRegionIDs {
	newRegionIDs := CachedRegionIDs{
		ECS: []string{},
		VPC: []string{},
	}
	ecsRegions := make(chan string, 50)
	vpcRegions := make(chan string, 50)
	var wg1, wg2 sync.WaitGroup
	wg1.Add(len(regionids.ECS))
	wg2.Add(len(regionids.VPC))
	for _, region := range regionids.ECS {
		go func(region string) {
			defer func() {
				wg1.Done()
			}()
			client, err := ecs.NewClientWithAccessKey(region, ctx.AccessKeyID, ctx.AccessSecret)
			if err != nil {
				logrus.Errorf("filter region failed, new client error:%s", err.Error())
				// ecsRegions <- region
			}

			request := ecs.CreateDescribeInstancesRequest()
			request.Scheme = "https"

			response, err := client.DescribeInstances(request)
			if err != nil {
				logrus.Errorf("filter region failed, describe instance error:%s", err.Error())
				// ecsRegions <- region
			}
			if len(response.Instances.Instance) > 0 {
				ecsRegions <- region
			}

		}(region)
	}

	for _, region := range regionids.VPC {
		go func(region string) {
			defer func() {
				wg2.Done()
			}()
			client, err := libvpc.NewClientWithAccessKey(region, ctx.AccessKeyID, ctx.AccessSecret)
			if err != nil {
				logrus.Errorf(err.Error())
				vpcRegions <- region
			}
			request := libvpc.CreateDescribeVpcsRequest()
			request.Scheme = "https"

			response, err := client.DescribeVpcs(request)
			if err != nil {
				logrus.Errorf(err.Error())
				vpcRegions <- region
			}
			if len(response.Vpcs.Vpc) > 0 {
				vpcRegions <- region
			}
		}(region)
	}
	wg1.Wait()
	close(ecsRegions)
	wg2.Wait()
	close(vpcRegions)
	for r := range ecsRegions {
		newRegionIDs.ECS = append(newRegionIDs.ECS, r)
	}
	for r := range vpcRegions {
		newRegionIDs.VPC = append(newRegionIDs.VPC, r)
	}

	return newRegionIDs
}

func GetCachedRegions(ak_ctx Context) *CachedRegionIDs {
	var regions CachedRegionIDs
	key := fmt.Sprintf("%s/%s/%s/%s", CloudResourcePrefix, ak_ctx.OrgID, ak_ctx.Vendor, ResourceRegions)
	err := ak_ctx.CachedJs.Get(context.Background(), key, &regions)
	if err != nil {
		logrus.Errorf("get cached cloud resource regions failed, key:%s, error:%v", key, err)
		return nil
	}
	if len(regions.VPC) == 0 {
		return nil
	}
	return &regions
}

func PutCachedRegions(ak_ctx Context, regions *CachedRegionIDs) {
	if regions == nil || len(regions.VPC) == 0 {
		return
	}
	now := time.Now()
	timestamp := now.Unix()
	regions.LastUpdateTimestamp = timestamp
	key := fmt.Sprintf("%s/%s/%s/%s", CloudResourcePrefix, ak_ctx.OrgID, ak_ctx.Vendor, ResourceRegions)
	err := ak_ctx.CachedJs.Put(context.Background(), key, *regions)
	if err != nil {
		logrus.Errorf("put cached cloud resource regions failed, key:%s, error:%v", key, err)
	}
}

func UpdateGlobalCachedRegions(regions *CachedRegionIDs) {
	if regions == nil || len(regions.VPC) == 0 {
		return
	}
	if ecsRegionIDs == nil {
		ecsRegionIDs = set.New()
	}
	if vpcRegionIDs == nil {
		vpcRegionIDs = set.New()
	}
	for _, region := range regions.VPC {
		if !vpcRegionIDs.Has(region) {
			vpcRegionIDs.Insert(region)
			globalRegionIDs.VPC = append(globalRegionIDs.VPC, region)
		}
	}
	for _, region := range regions.ECS {
		if !ecsRegionIDs.Has(region) {
			ecsRegionIDs.Insert(region)
			globalRegionIDs.VPC = append(globalRegionIDs.VPC, region)
		}
	}
}
