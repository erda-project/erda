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

package crondsvc

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *CrondSvc) CleanBuildCacheImages() {
	alertErr := func(err error) {
		logrus.Errorf("[alert] failed to clean build cache images, err: %v", err)
	}

	alertErrWithCluster := func(err error, clusterName string) {
		logrus.Errorf("[alert] failed to clean build cache images, clusterName: %s, err: %v", clusterName, err)
	}
	//假如时间多少天都没有变更
	date := time.Now().Add(-conf.BuildCacheExpireIn())

	var toDeleteCacheImages []spec.CIV3BuildCache
	if err := s.dbClient.Where("last_pull_at is null and created_at < ?", date).Or("last_pull_at < ?", date).
		Find(&toDeleteCacheImages); err != nil {
		alertErr(err)
		return
	}

	if len(toDeleteCacheImages) == 0 {
		return
	}

	imageMap := make(map[string][]spec.CIV3BuildCache, 0)
	for _, v := range toDeleteCacheImages {
		images, ok := imageMap[v.ClusterName]
		if ok {
			images = append(images, v)
			imageMap[v.ClusterName] = images
		} else {
			imageMap[v.ClusterName] = []spec.CIV3BuildCache{v}
		}
	}

	for clusterName, images := range imageMap {
		var imageNames []string
		for _, v := range images {
			imageNames = append(imageNames, v.Name)
		}
		result, err := s.bdl.DeleteImageManifests(clusterName, imageNames)
		if err != nil {
			alertErrWithCluster(err, clusterName)
			continue
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			alertErrWithCluster(err, clusterName)
			continue
		}

		logrus.Errorf("[alert] clusterName: %s, delete build cache success: %s", clusterName, string(bytes))

		for _, name := range result.Succeed {
			cache, err := s.dbClient.GetBuildCache(clusterName, name)
			if err != nil {
				alertErrWithCluster(err, clusterName)
				continue
			}
			if err = s.dbClient.DeleteBuildCache(cache.ID); err != nil {
				alertErrWithCluster(err, clusterName)
				continue
			}
		}
	}
}
