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

package db

import (
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
)

type CIV3BuildCache struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`
	Name        string    `json:"name"`
	ClusterName string    `json:"clusterName"`
	LastPullAt  time.Time `json:"lastPullAt"`
	CreatedAt   time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt   time.Time `json:"updatedAt" xorm:"updated"`
	DeletedAt   time.Time `xorm:"deleted"`
}

func (*CIV3BuildCache) TableName() string {
	return "ci_v3_build_caches"
}

func (client *Client) GetBuildCache(clusterName, imageName string, ops ...mysqlxorm.SessionOption) (cache CIV3BuildCache, err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	defer func() {
		err = errors.Wrapf(err, "failed to get build cache, clusterName [%s], imageName [%s]", clusterName, imageName)
	}()

	cache.ClusterName = clusterName
	cache.Name = imageName
	ok, err := session.Get(&cache)
	if err != nil {
		return CIV3BuildCache{}, err
	}
	if !ok {
		return CIV3BuildCache{}, errors.New("not found")
	}
	return cache, nil
}

func (client *Client) DeleteBuildCache(id int64, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	defer func() { err = errors.Wrapf(err, "failed to delete build cache, id [%v]", id) }()

	_, err = session.ID(id).Delete(&CIV3BuildCache{})
	return err
}
