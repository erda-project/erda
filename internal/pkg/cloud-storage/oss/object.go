// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

	"github.com/erda-project/erda/internal/pkg/cloud-storage/types"
)

func (p *provider) WhoIAm() string {
	return "oss"
}

func (p *provider) PutObject(ctx context.Context, localFilepath string, filename string) (*types.PutObjectResult, error) {
	key := fmt.Sprintf("%s/%s", p.Cfg.Prefix, filename)
	request := oss.PutObjectRequest{
		Bucket:       &p.Cfg.Bucket,
		Key:          &key,
		StorageClass: oss.StorageClassStandard,
		Acl:          oss.ObjectACLPublicRead,
	}
	result, err := p.client.PutObjectFromFile(ctx, &request, localFilepath)
	if err != nil {
		return nil, err
	}
	return p.buildResult(result, key), nil
}

func (p *provider) PutObjectWithReader(ctx context.Context, reader io.Reader, filename string) (*types.PutObjectResult, error) {
	key := fmt.Sprintf("%s/%s", p.Cfg.Prefix, filename)
	request := oss.PutObjectRequest{
		Bucket: &p.Cfg.Bucket,
		Key:    &key,
		Body:   reader,
		Acl:    oss.ObjectACLPublicRead,
	}
	result, err := p.client.PutObject(ctx, &request)
	if err != nil {
		return nil, err
	}
	return p.buildResult(result, key), nil
}

func (p *provider) buildResult(result *oss.PutObjectResult, objName string) *types.PutObjectResult {
	var res types.PutObjectResult
	if result.ETag != nil {
		res.ETag = *result.ETag
	}
	if result.ContentMD5 != nil {
		res.ContentMD5 = *result.ContentMD5
	}
	if result.HashCRC64 != nil {
		res.HashCRC64 = *result.HashCRC64
	}
	if result.VersionId != nil {
		res.VersionId = *result.VersionId
	}
	res.Bucket = p.Cfg.Bucket
	res.Region = p.Cfg.Region
	res.ObjectName = objName
	return &res
}

func (p *provider) ListObjects(ctx context.Context) ([]types.FileObject, error) {
	maxKey := p.Cfg.MaxKey
	if maxKey <= 0 {
		maxKey = 1000
	}
	var res []types.FileObject
	result, err := p.listObjects(ctx, maxKey)
	if err != nil {
		return nil, err
	}
	for _, content := range result {
		res = append(res, types.FileObject{
			Key:          *content.Key,
			Type:         content.Type,
			Size:         content.Size,
			LastModified: content.LastModified,
		})
	}
	return res, nil
}

func (p *provider) listObjects(ctx context.Context, maxKeys int32) ([]oss.ObjectProperties, error) {
	if maxKeys == 0 {
		maxKeys = 1000
	}
	var res []oss.ObjectProperties
	object, err := p.client.ListObjects(ctx, &oss.ListObjectsRequest{
		Bucket:  &p.Cfg.Bucket,
		MaxKeys: maxKeys,
		Prefix:  &p.Cfg.Prefix,
	})
	if err != nil {
		return nil, err
	}
	res = append([]oss.ObjectProperties{}, object.Contents...)
	if object.IsTruncated {
		result, err := p.listObjects(ctx, maxKeys+maxKeys)
		if err != nil {
			return nil, err
		}
		res = append([]oss.ObjectProperties{}, result...)
		return res, nil
	}
	return res, nil
}

func (p *provider) DeleteObject(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: &p.Cfg.Bucket,
		Key:    &key,
	})
	return err
}

func (p *provider) DeleteObjects(ctx context.Context, keys []string) error {
	var deleteObjects []oss.DeleteObject
	for _, key := range keys {
		deleteObjects = append(deleteObjects, oss.DeleteObject{
			Key: &key,
		})
	}
	_, err := p.client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket:  &p.Cfg.Bucket,
		Objects: deleteObjects,
	})
	return err
}
