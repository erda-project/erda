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
