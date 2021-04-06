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

package clusterinfo

import (
	"context"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/strutil"
)

// clusterInfoPrefix 是集群配置信息在 etcd 中的路径前缀
const (
	clusterInfoPrefix = "/dice/scheduler/clusterinfo/"
	queryTimeout      = 3 * time.Second
)

type ClusterInfo interface {
	Info(string) (apistructs.ClusterInfoData, error)
	List([]string) (apistructs.ClusterInfoDataList, error)
}

type ClusterInfoImpl struct {
	js jsonstore.JsonStore
}

func NewClusterInfoImpl(js jsonstore.JsonStore) ClusterInfo {
	return &ClusterInfoImpl{js: js}
}

func (c *ClusterInfoImpl) Info(name string) (apistructs.ClusterInfoData, error) {
	var data apistructs.ClusterInfoData

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	if err := c.js.Get(ctx, strutil.Concat(clusterInfoPrefix, name), &data); err != nil {
		return apistructs.ClusterInfoData{}, err
	}

	return data, nil
}

func (c *ClusterInfoImpl) List(names []string) (apistructs.ClusterInfoDataList, error) {
	var dataList apistructs.ClusterInfoDataList
	var nameMap map[string]interface{}

	for _, name := range names {
		nameMap[name] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := c.js.ForEach(ctx, clusterInfoPrefix, apistructs.ClusterInfoData{}, func(key string, v interface{}) error {
		if _, ok := nameMap[key]; ok {
			data := v.(*apistructs.ClusterInfoData)
			dataList = append(dataList, *data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dataList, nil
}
