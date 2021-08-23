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

package buildcachesvc

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type BuildCacheSvc struct {
	dbClient *dbclient.Client
}

func New(dbClient *dbclient.Client) *BuildCacheSvc {
	s := BuildCacheSvc{}
	s.dbClient = dbClient
	return &s
}

func (s *BuildCacheSvc) Report(req *apistructs.BuildCacheImageReportRequest, cache *spec.CIV3BuildCache) error {
	success, err := s.dbClient.Get(cache)
	if err != nil {
		return apierrors.ErrReportBuildCache.InternalError(err)
	}
	if req.Action == "push" {
		// 不存在添加,存在不处理
		if !success {
			if _, err = s.dbClient.Insert(cache); err != nil {
				return apierrors.ErrReportBuildCache.InternalError(err)
			}
		}

	} else if req.Action == "pull" {
		// 存在更新时间,不存在不处理
		if success {
			cache.LastPullAt = time.Now()
			if _, err = s.dbClient.ID(cache.ID).Update(cache); err != nil {
				return apierrors.ErrReportBuildCache.InternalError(err)
			}
		}
	}

	return nil
}
