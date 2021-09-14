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

package authentication

import (
	"context"
	"sync"

	akpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
)

type AccessItemCollection map[string]*akpb.AccessKeysItem

type Validator interface {
	// Validate +Validate
	Validate(scope string, scopeId string, accessKeyId string, accessKeySecret string) bool
}

type accessKeyValidator struct {
	sync.RWMutex
	collection       AccessItemCollection
	AccessKeyService akpb.AccessKeyServiceServer
}

func (v *accessKeyValidator) syncFullAccessKeys(ctx context.Context) error {
	var (
		pageNumber int64 = 1
		pageSize   int64 = 100
	)
	results := make([]*akpb.AccessKeysItem, 0)
	for {
		resp, err := v.AccessKeyService.QueryAccessKeys(ctx, &akpb.QueryAccessKeysRequest{
			PageNo:   pageNumber,
			PageSize: pageSize,
		})
		if err != nil {
			return err
		}
		if resp.Data != nil {
			results = append(results, resp.Data...)
		}
		if resp.Total <= pageNumber*pageSize {
			break
		}
		pageNumber++
	}
	v.Lock()
	for k, _ := range v.collection {
		delete(v.collection, k)
	}
	for _, item := range results {
		v.collection[item.AccessKey] = item
	}
	v.Unlock()
	return nil
}

func (v *accessKeyValidator) Validate(scope string, scopeId string, accessKeyId string, accessKeySecret string) bool {
	v.RLock()
	defer v.RUnlock()
	item, ok := v.collection[accessKeyId]
	return ok && item.AccessKey == accessKeyId && item.SecretKey == accessKeySecret && item.ScopeId == scopeId && item.Scope == scope
}
